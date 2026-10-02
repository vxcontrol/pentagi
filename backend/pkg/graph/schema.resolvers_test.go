package graph

import (
	"cmp"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/database/knowledge"
	"pentagi/pkg/flowfiles"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// graphResolver wires what most mutations reach for; a test sets the collaborators its resolver adds.
func graphResolver(db *graphDB) *Resolver {
	return &Resolver{
		DB:            db,
		Config:        &config.Config{CookieSigningSalt: "signing-salt"},
		Logger:        logrus.NewEntry(logrus.New()),
		Subscriptions: subscriptions.NewSubscriptionsController(),
	}
}

// graphSessionContext is user 1 on a browser session, the only kind the token and template resolvers accept.
func graphSessionContext(perms ...string) context.Context {
	return SetUserType(graphUserContext(1, perms...), "local")
}

type graphFlowCtrl struct {
	controller.FlowController

	renameErr       error
	finished        []int64
	providerRenames []provider.ProviderName
}

func (c *graphFlowCtrl) FinishFlow(ctx context.Context, flowID int64) error {
	return graphWrite(ctx, func() { c.finished = append(c.finished, flowID) })
}

func (c *graphFlowCtrl) RenameFlow(ctx context.Context, _ int64, _ string) error {
	return cmp.Or(ctx.Err(), c.renameErr)
}

func (c *graphFlowCtrl) RenameFlowsProvider(ctx context.Context, _ int64, _, newName provider.ProviderName) error {
	return graphWrite(ctx, func() { c.providerRenames = append(c.providerRenames, newName) })
}

type graphProvidersCtrl struct {
	providers.ProviderController
}

func (graphProvidersCtrl) UpdateProvider(
	ctx context.Context, _ int64, prvID int64, prvname provider.ProviderName, _ *pconfig.ProviderConfig,
) (database.Provider, error) {
	return database.Provider{ID: prvID, Name: strings.TrimSpace(string(prvname))}, ctx.Err()
}

func (graphProvidersCtrl) SeedDefaultProviders(ctx context.Context, _ int64) error { return ctx.Err() }

func (graphProvidersCtrl) DefaultProvidersConfig() provider.ProvidersConfig {
	return provider.ProvidersConfig{provider.ProviderGLM: &pconfig.ProviderConfig{}}
}

func (graphProvidersCtrl) DefaultProviders() provider.Providers { return provider.Providers{} }

type graphKnowledgeStore struct {
	knowledge.KnowledgeStore

	created []model.CreateKnowledgeDocumentInput
	updated []model.UpdateKnowledgeDocumentInput
	renamed []string
}

func (s *graphKnowledgeStore) CreateDocument(
	ctx context.Context, _ int64, input model.CreateKnowledgeDocumentInput,
) (*model.KnowledgeDocument, error) {
	err := graphWrite(ctx, func() { s.created = append(s.created, input) })
	return &model.KnowledgeDocument{ID: "doc-1", Question: input.Question}, err
}

func (s *graphKnowledgeStore) UpdateUserDocument(
	ctx context.Context, _ int64, id string, input model.UpdateKnowledgeDocumentInput,
) (*model.KnowledgeDocument, error) {
	err := graphWrite(ctx, func() { s.updated = append(s.updated, input) })
	return &model.KnowledgeDocument{ID: id}, err
}

func (s *graphKnowledgeStore) RenameUserDocument(
	ctx context.Context, _ int64, _, question string,
) (*model.KnowledgeDocument, error) {
	err := graphWrite(ctx, func() { s.renamed = append(s.renamed, question) })
	return &model.KnowledgeDocument{ID: "doc-1", Question: question}, err
}

func TestSchemaResolvers_DeleteFlow_FinishesTheWorkerAndDropsItFromTheOwnersFavorites(t *testing.T) {
	db := &graphDB{flowOwner: 7}
	flows := &graphFlowCtrl{}
	r := graphResolver(db)
	r.Controller = flows

	ownerFrames, err := r.Subscriptions.NewSettingsSubscriber(7).SettingsUserUpdated(t.Context())
	require.NoError(t, err)
	callerFrames, err := r.Subscriptions.NewSettingsSubscriber(9).SettingsUserUpdated(t.Context())
	require.NoError(t, err)

	// An admin deleting someone else's flow is what tells the owner's id apart from the caller's.
	_, err = r.Mutation().DeleteFlow(graphUserContext(9, "flows.admin"), 42)
	require.NoError(t, err)

	assert.Equal(t, []int64{42}, flows.finished, "the worker of a deleted flow stays in memory for the life of the process")
	assert.Equal(t, []database.DeleteFavoriteFlowParams{{FlowID: 42, UserID: 7}}, db.droppedFavorites,
		"the id of a deleted flow stays starred")
	select {
	case <-ownerFrames:
	default:
		t.Error("favorites changed without SettingsUserUpdated to the owner, so a connected client keeps the dead id")
	}
	select {
	case <-callerFrames:
		t.Error("SettingsUserUpdated went to the caller, who gets someone else's preferences")
	default:
	}
}

func TestSchemaResolvers_RenameFlow_FallsBackForAFlowWithNoWorker(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "the row is gone", err: controller.ErrFlowNotFound},
		{name: "the row is there but nothing loaded it", err: controller.ErrFlowNotLoaded},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := &graphDB{flowOwner: 1}
			r := graphResolver(db)
			r.Controller = &graphFlowCtrl{renameErr: tt.err}

			result, err := r.Mutation().RenameFlow(graphUserContext(1, "flows.edit"), 7, "new title")

			require.NoError(t, err, "the title the user typed was dropped")
			assert.Equal(t, model.ResultTypeSuccess, result)
			assert.Equal(t, []string{"new title"}, db.flowTitles)
		})
	}
}

func TestSchemaResolvers_UpdateProvider_CascadesTheStoredName(t *testing.T) {
	for _, tt := range []struct {
		name        string
		argument    string
		wantCascade []provider.ProviderName
	}{
		{name: "plain rename", argument: "B", wantCascade: []provider.ProviderName{"B"}},
		{name: "padded rename cascades the trimmed name", argument: "  B  ", wantCascade: []provider.ProviderName{"B"}},
		{name: "padding alone is not a rename", argument: "  A  "},
	} {
		t.Run(tt.name, func(t *testing.T) {
			flows := &graphFlowCtrl{}
			r := graphResolver(&graphDB{provider: database.Provider{ID: 7, Name: "A"}})
			r.ProvidersCtrl = graphProvidersCtrl{}
			r.Controller = flows

			_, err := r.Mutation().UpdateProvider(graphUserContext(1, "settings.providers.edit"), 7, tt.argument, model.AgentsConfig{})

			require.NoError(t, err)
			assert.Equal(t, tt.wantCascade, flows.providerRenames, "flows repointed to a name no provider row holds")
		})
	}
}

func TestSchemaResolvers_DeletePrompt_ReportsAMissingRow(t *testing.T) {
	for _, tt := range []struct {
		name        string
		found       bool
		wantResult  model.ResultType
		wantErr     error
		wantDeletes int
	}{
		{name: "own prompt is deleted", found: true, wantResult: model.ResultTypeSuccess, wantDeletes: 1},
		{name: "missing or foreign id is an error", wantResult: model.ResultTypeError, wantErr: sql.ErrNoRows},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := &graphDB{promptFound: tt.found}

			result, err := graphResolver(db).Mutation().DeletePrompt(graphUserContext(1, "settings.prompts.edit"), 1)

			assert.Equal(t, tt.wantResult, result)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantDeletes, db.promptDeletes)
		})
	}
}

func TestSchemaResolvers_CreateAPIToken_MeasuresTheNameInCharacters(t *testing.T) {
	for _, tt := range []struct {
		name      string
		tokenName string
		wantErr   string
	}{
		{name: "a hundred Cyrillic characters fit the limit", tokenName: strings.Repeat("я", 100)},
		{
			name:      "one Cyrillic character over the limit is refused",
			tokenName: strings.Repeat("я", 101),
			wantErr:   "token name must not exceed 100 characters",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := &graphDB{}
			input := model.CreateAPITokenInput{Name: &tt.tokenName, TTL: 3600}

			_, err := graphResolver(db).Mutation().CreateAPIToken(graphSessionContext("settings.tokens.create"), input)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Empty(t, db.createdTokens, "a refused name reached the database")
				return
			}
			require.ErrorIs(t, err, errStoppedBeforeTokenCache, "the resolver never reached the database")
			require.Len(t, db.createdTokens, 1)
			assert.Equal(t, sql.NullString{String: tt.tokenName, Valid: true}, db.createdTokens[0].Name)
		})
	}
}

func TestSchemaResolvers_UpdateAPIToken_StoresTheValidatedNameAndStatus(t *testing.T) {
	status := func(s model.TokenStatus) *model.TokenStatus { return &s }

	for _, tt := range []struct {
		name     string
		input    model.UpdateAPITokenInput
		wantErr  string
		wantName sql.NullString
	}{
		{
			name:     "a hundred Cyrillic characters fit the limit",
			input:    model.UpdateAPITokenInput{Name: strPtr(strings.Repeat("я", 100))},
			wantName: sql.NullString{String: strings.Repeat("я", 100), Valid: true},
		},
		{
			name:    "one Cyrillic character over the limit is refused",
			input:   model.UpdateAPITokenInput{Name: strPtr(strings.Repeat("я", 101))},
			wantErr: "token name must not exceed 100 characters",
		},
		{name: "an empty name clears the stored one", input: model.UpdateAPITokenInput{Name: strPtr("")}},
		{
			name:     "an omitted name keeps the stored one",
			input:    model.UpdateAPITokenInput{Status: status(model.TokenStatusActive)},
			wantName: sql.NullString{String: "stored name", Valid: true},
		},
		{
			name:     "the derived expired status keeps the stored one",
			input:    model.UpdateAPITokenInput{Name: strPtr("renamed"), Status: status(model.TokenStatusExpired)},
			wantName: sql.NullString{String: "renamed", Valid: true},
		},
		{
			name:    "a status outside the enum is refused",
			input:   model.UpdateAPITokenInput{Status: status("bogus")},
			wantErr: "invalid token status: bogus",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := &graphDB{}

			_, err := graphResolver(db).Mutation().UpdateAPIToken(graphSessionContext("settings.tokens.edit"), "tok-1", tt.input)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Empty(t, db.updatedTokens, "a refused update reached the database")
				return
			}
			require.NoError(t, err)
			require.Len(t, db.updatedTokens, 1)
			assert.Equal(t, tt.wantName, db.updatedTokens[0].Name)
			assert.Equal(t, database.TokenStatusActive, db.updatedTokens[0].Status, "the stored status changed")
		})
	}
}

func TestSchemaResolvers_AddFavoriteFlow_AcceptsOnlyYourOwnFlow(t *testing.T) {
	for _, tt := range []struct {
		name          string
		owner         int64
		perm          string
		wantResult    model.ResultType
		wantErr       error
		wantFavorites []database.AddFavoriteFlowParams
	}{
		{
			name:       "another user's flow is refused even to an admin",
			owner:      2,
			perm:       "flows.admin",
			wantResult: model.ResultTypeError,
			wantErr:    ErrForbidden,
		},
		{
			name:          "your own flow is starred",
			owner:         1,
			perm:          "flows.view",
			wantResult:    model.ResultTypeSuccess,
			wantFavorites: []database.AddFavoriteFlowParams{{UserID: 1, FlowID: 5}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db := &graphDB{flowOwner: tt.owner}

			result, err := graphResolver(db).Mutation().AddFavoriteFlow(graphSessionContext(tt.perm, "settings.user.edit"), 5)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantResult, result)
			assert.Equal(t, tt.wantFavorites, db.favorites)
		})
	}
}

// graphFlowTemplateCases are the entry's two outcomes; validateFlowTemplateFields has the field classes.
var graphFlowTemplateCases = []struct {
	name      string
	title     string
	text      string
	wantTitle string
	wantText  string
	wantErr   string
}{
	{
		name:      "the title is trimmed and the text is written as it was sent",
		title:     "  My template  ",
		text:      "\n  scan the host  \n",
		wantTitle: "My template",
		wantText:  "\n  scan the host  \n",
	},
	{name: "a blank text is refused before the row is written", title: "My template", text: " \n\t", wantErr: "text is required"},
	{name: "a blank title is refused before the row is written", title: "   ", text: "scan the host", wantErr: "title is required"},
}

func TestSchemaResolvers_CreateFlowTemplate_StoresTheValidatedFields(t *testing.T) {
	for _, tt := range graphFlowTemplateCases {
		t.Run(tt.name, func(t *testing.T) {
			db := &graphDB{}
			input := model.CreateFlowTemplateInput{Title: tt.title, Text: tt.text}

			template, err := graphResolver(db).Mutation().CreateFlowTemplate(graphSessionContext("templates.create"), input)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Empty(t, db.createdTemplates, "a refused template reached the database")
				return
			}
			require.NoError(t, err)
			require.Len(t, db.createdTemplates, 1)
			assert.Equal(t, tt.wantTitle, db.createdTemplates[0].Title)
			assert.Equal(t, tt.wantText, db.createdTemplates[0].Text)
			assert.Equal(t, tt.wantTitle, template.Title, "the returned title is one no template row holds")
		})
	}
}

func TestSchemaResolvers_UpdateFlowTemplate_StoresTheValidatedFields(t *testing.T) {
	for _, tt := range graphFlowTemplateCases {
		t.Run(tt.name, func(t *testing.T) {
			db := &graphDB{}
			input := model.UpdateFlowTemplateInput{Title: tt.title, Text: tt.text}

			template, err := graphResolver(db).Mutation().UpdateFlowTemplate(graphSessionContext("templates.edit"), 1, input)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Empty(t, db.updatedTemplates, "a refused template reached the database")
				return
			}
			require.NoError(t, err)
			require.Len(t, db.updatedTemplates, 1)
			assert.Equal(t, tt.wantTitle, db.updatedTemplates[0].Title)
			assert.Equal(t, tt.wantText, db.updatedTemplates[0].Text)
			assert.Equal(t, tt.wantTitle, template.Title, "the returned title is one no template row holds")
		})
	}
}

// Subtests are keyed by resolver.
func TestSchemaResolvers_AMissingFlowTemplateIsNotFound(t *testing.T) {
	r := graphResolver(&graphDB{templateMissing: true})

	for _, tt := range []struct {
		name string
		call func() error
	}{
		{name: "reading it", call: func() error {
			_, err := r.Query().FlowTemplate(graphSessionContext("templates.view"), 99)
			return err
		}},
		{name: "updating it", call: func() error {
			input := model.UpdateFlowTemplateInput{Title: "t", Text: "x"}
			_, err := r.Mutation().UpdateFlowTemplate(graphSessionContext("templates.edit"), 99, input)
			return err
		}},
		{name: "deleting it", call: func() error {
			_, err := r.Mutation().DeleteFlowTemplate(graphSessionContext("templates.delete"), 99)
			return err
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, tt.call(), sql.ErrNoRows, "the presenter classifies only sql.ErrNoRows as NOT_FOUND")
		})
	}
}

func TestSchemaResolvers_CreateKnowledgeDocument_RefusesABlankQuestionOrContent(t *testing.T) {
	for _, tt := range []struct {
		name     string
		question string
		content  string
		wantErr  string
	}{
		{name: "an ordinary question is stored", question: "which ports are open?", content: "nmap reports 22 and 443"},
		{name: "a blank question is refused", question: "   ", content: "nmap reports 22 and 443", wantErr: "question is required"},
		{name: "a blank content is refused", question: "which ports are open?", content: " \n\t", wantErr: "content is required"},
		{name: "a padded content reaches the store as it was sent", question: "which ports are open?", content: "  22\n443\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &graphKnowledgeStore{}
			r := graphResolver(&graphDB{})
			r.Knowledge = store
			input := model.CreateKnowledgeDocumentInput{
				DocType:  model.KnowledgeDocTypeAnswer,
				Content:  tt.content,
				Question: tt.question,
			}

			_, err := r.Mutation().CreateKnowledgeDocument(graphUserContext(1, "knowledge.create"), input)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Empty(t, store.created, "a refused document reached the store")
				return
			}
			require.NoError(t, err)
			require.Len(t, store.created, 1)
			assert.Equal(t, tt.question, store.created[0].Question)
			assert.Equal(t, tt.content, store.created[0].Content)
		})
	}
}

func TestSchemaResolvers_UpdateKnowledgeDocument_RefusesABlankQuestionOrContent(t *testing.T) {
	for _, tt := range []struct {
		name     string
		question *string
		content  string
		wantErr  string
	}{
		{name: "an ordinary question is stored", question: strPtr("which ports are open?"), content: "nmap reports 22 and 443"},
		{name: "an omitted question leaves the stored one alone", question: nil, content: "nmap reports 22 and 443"},
		{name: "a blank question is refused", question: strPtr("   "), content: "nmap reports 22 and 443", wantErr: "question is required"},
		{name: "a blank content is refused", question: nil, content: " \n\t", wantErr: "content is required"},
		{name: "a padded content reaches the store as it was sent", question: nil, content: "  22\n443\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &graphKnowledgeStore{}
			r := graphResolver(&graphDB{})
			r.Knowledge = store
			input := model.UpdateKnowledgeDocumentInput{Content: tt.content, Question: tt.question}

			_, err := r.Mutation().UpdateKnowledgeDocument(graphUserContext(1, "knowledge.edit"), "doc-1", input)

			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				assert.Empty(t, store.updated, "a refused document reached the store")
				return
			}
			require.NoError(t, err)
			require.Len(t, store.updated, 1)
			assert.Equal(t, tt.question, store.updated[0].Question)
			assert.Equal(t, tt.content, store.updated[0].Content)
		})
	}
}

func TestSchemaResolvers_RenameKnowledgeDocument_RefusesAnInvalidQuestion(t *testing.T) {
	for _, tt := range []struct {
		name     string
		question string
		wantErr  string
	}{
		{name: "an ordinary question is stored", question: "how do I check the kernel version?"},
		{name: "a multibyte question at the length limit is stored", question: strings.Repeat("я", maxKnowledgeQuestionLen)},
		{name: "a blank question is refused", question: "   ", wantErr: "question is required"},
		{
			name:     "a question one over the length limit is refused",
			question: strings.Repeat("a", maxKnowledgeQuestionLen+1),
			wantErr:  "question must not exceed",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			store := &graphKnowledgeStore{}
			r := graphResolver(&graphDB{})
			r.Knowledge = store

			_, err := r.Mutation().RenameKnowledgeDocument(graphUserContext(1, "knowledge.edit"), "doc-1", tt.question)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Empty(t, store.renamed, "a refused rename reached the store")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{tt.question}, store.renamed)
		})
	}
}

func TestSchemaResolvers_FlowFiles_CarryTheFlowTheyWereAskedFor(t *testing.T) {
	dataDir := t.TempDir()
	uploads := flowfiles.FlowUploadsDir(dataDir, 42)
	require.NoError(t, os.MkdirAll(uploads, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(uploads, "report.txt"), []byte("body"), 0o644))

	r := graphResolver(&graphDB{flowOwner: 1})
	r.Config.DataDir = dataDir

	files, err := r.Query().FlowFiles(graphUserContext(1, "flow_files.view"), 42)

	require.NoError(t, err)
	require.Len(t, files, 1, "want the one file that was seeded")
	assert.Equal(t, int64(42), files[0].FlowID)
}

func TestSchemaResolvers_Resources_PrefersTheCallersOwnRowOnPathCollision(t *testing.T) {
	// The administrator branch drops the owner predicate, so both rows of a colliding path come back.
	db := &graphDB{resources: []database.UserResource{
		{ID: 10, UserID: 2, Name: "report.txt", Path: "report.txt"},
		{ID: 11, UserID: 1, Name: "report.txt", Path: "report.txt"},
	}}

	got, err := graphResolver(db).Query().Resources(graphUserContext(1, "resources.admin"), nil, nil)

	require.NoError(t, err)
	require.Len(t, got, 1, "one row per path")
	assert.Equal(t, int64(1), got[0].UserID, "want the caller's own row")
}

func TestSchemaResolvers_SettingsProviders_ReadsTheThinkingBudgetRuleOnTheConfiguredPrefix(t *testing.T) {
	for _, tt := range []struct {
		name    string
		prefix  string
		refused bool
	}{
		{name: "the zai gateway takes no thinking depth", prefix: "zai", refused: true},
		{name: "the dashscope gateway takes one", prefix: "dashscope", refused: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := graphResolver(&graphDB{})
			r.Config.GLMProvider = tt.prefix
			r.ProvidersCtrl = graphProvidersCtrl{}

			got, err := r.Query().SettingsProviders(graphUserContext(1, "settings.providers.view"))
			require.NoError(t, err)

			for _, m := range got.Models.Glm {
				if m.Name == "glm-5" {
					require.NotNil(t, m.Reasoning)
					require.NotNil(t, m.Reasoning.TakesNoThinkingDepth)
					assert.Equal(t, tt.refused, *m.Reasoning.TakesNoThinkingDepth)
					return
				}
			}
			t.Fatal("glm-5 is missing from the GLM catalogue")
		})
	}
}
