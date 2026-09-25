package graph

import (
	"cmp"
	"context"
	"database/sql"
	"errors"

	"pentagi/pkg/database"
)

// errStoppedBeforeTokenCache ends CreateAPIToken at its insert: the token cache after it needs gorm.
var errStoppedBeforeTokenCache = errors.New("the token cache is not wired in this test")

// graphWrite records a write only on a live context, the way the real write fails on a done one.
func graphWrite(ctx context.Context, record func()) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	record()
	return nil
}

// graphDB records the writes a test asserts on, so a test tells a refused mutation from one that reached the database.
type graphDB struct {
	database.Querier

	flowOwner       int64
	resources       []database.UserResource
	provider        database.Provider
	promptFound     bool
	templateMissing bool

	promptDeletes    int
	createdTokens    []database.CreateAPITokenParams
	updatedTokens    []database.UpdateUserAPITokenParams
	createdTemplates []database.CreateFlowTemplateParams
	updatedTemplates []database.UpdateFlowTemplateParams
	favorites        []database.AddFavoriteFlowParams
	droppedFavorites []database.DeleteFavoriteFlowParams
	flowTitles       []string
}

func (db *graphDB) GetFlow(_ context.Context, id int64) (database.Flow, error) {
	return database.Flow{ID: id, UserID: db.flowOwner}, nil
}

func (db *graphDB) GetFlowContainers(context.Context, int64) ([]database.Container, error) {
	return nil, nil
}

func (db *graphDB) GetUserResourcesByIDs(context.Context, []int64) ([]database.UserResource, error) {
	return db.resources, nil
}

func (db *graphDB) GetAllResourcesRoot(context.Context) ([]database.UserResource, error) {
	return db.resources, nil
}

func (db *graphDB) GetUserProvider(context.Context, database.GetUserProviderParams) (database.Provider, error) {
	return db.provider, nil
}

func (db *graphDB) GetUserProviders(context.Context, int64) ([]database.Provider, error) {
	return nil, nil
}

func (db *graphDB) GetUser(_ context.Context, id int64) (database.GetUserRow, error) {
	return database.GetUserRow{ID: id, Hash: "user-hash", RoleID: 2}, nil
}

func (db *graphDB) GetUserPrompt(context.Context, database.GetUserPromptParams) (database.Prompt, error) {
	if !db.promptFound {
		return database.Prompt{}, sql.ErrNoRows
	}
	return database.Prompt{ID: 1, UserID: 1}, nil
}

func (db *graphDB) DeleteUserPrompt(ctx context.Context, _ database.DeleteUserPromptParams) error {
	return graphWrite(ctx, func() { db.promptDeletes++ })
}

func (db *graphDB) CreateAPIToken(ctx context.Context, arg database.CreateAPITokenParams) (database.ApiToken, error) {
	err := graphWrite(ctx, func() { db.createdTokens = append(db.createdTokens, arg) })
	return database.ApiToken{}, cmp.Or(err, errStoppedBeforeTokenCache)
}

func (db *graphDB) GetUserAPITokenByTokenID(
	_ context.Context, arg database.GetUserAPITokenByTokenIDParams,
) (database.ApiToken, error) {
	return database.ApiToken{
		ID:      5,
		TokenID: arg.TokenID,
		UserID:  arg.UserID,
		Name:    sql.NullString{String: "stored name", Valid: true},
		Status:  database.TokenStatusActive,
	}, nil
}

func (db *graphDB) UpdateUserAPIToken(ctx context.Context, arg database.UpdateUserAPITokenParams) (database.ApiToken, error) {
	err := graphWrite(ctx, func() { db.updatedTokens = append(db.updatedTokens, arg) })
	return database.ApiToken{ID: arg.ID, UserID: arg.UserID, Name: arg.Name, Status: arg.Status}, err
}

func (db *graphDB) CreateFlowTemplate(
	ctx context.Context, arg database.CreateFlowTemplateParams,
) (database.FlowTemplate, error) {
	err := graphWrite(ctx, func() { db.createdTemplates = append(db.createdTemplates, arg) })
	return database.FlowTemplate{ID: 1, UserID: arg.UserID, Title: arg.Title, Text: arg.Text}, err
}

func (db *graphDB) GetFlowTemplate(_ context.Context, arg database.GetFlowTemplateParams) (database.FlowTemplate, error) {
	if db.templateMissing {
		return database.FlowTemplate{}, sql.ErrNoRows
	}
	return database.FlowTemplate{ID: arg.ID, UserID: arg.UserID, Title: "stored title", Text: "stored text"}, nil
}

func (db *graphDB) UpdateFlowTemplate(
	ctx context.Context, arg database.UpdateFlowTemplateParams,
) (database.FlowTemplate, error) {
	err := graphWrite(ctx, func() { db.updatedTemplates = append(db.updatedTemplates, arg) })
	return database.FlowTemplate{ID: arg.ID, UserID: arg.UserID, Title: arg.Title, Text: arg.Text}, err
}

func (db *graphDB) AddFavoriteFlow(ctx context.Context, arg database.AddFavoriteFlowParams) (database.UserPreference, error) {
	err := graphWrite(ctx, func() { db.favorites = append(db.favorites, arg) })
	return database.UserPreference{UserID: arg.UserID}, err
}

func (db *graphDB) DeleteFlow(ctx context.Context, id int64) (database.Flow, error) {
	return database.Flow{ID: id, UserID: db.flowOwner}, ctx.Err()
}

func (db *graphDB) DeleteFlowMemoryDocuments(ctx context.Context, _ sql.NullString) error {
	return ctx.Err()
}

func (db *graphDB) DeleteFavoriteFlow(
	ctx context.Context, arg database.DeleteFavoriteFlowParams,
) (database.UserPreference, error) {
	err := graphWrite(ctx, func() { db.droppedFavorites = append(db.droppedFavorites, arg) })
	return database.UserPreference{UserID: arg.UserID}, err
}

func (db *graphDB) UpdateFlowTitle(ctx context.Context, arg database.UpdateFlowTitleParams) (database.Flow, error) {
	err := graphWrite(ctx, func() { db.flowTitles = append(db.flowTitles, arg.Title) })
	return database.Flow{ID: arg.ID, UserID: db.flowOwner, Title: arg.Title}, err
}

func graphUserContext(uid uint64, perms ...string) context.Context {
	return SetUserPermissions(SetUserID(context.Background(), uid), perms)
}

func strPtr(s string) *string {
	return &s
}
