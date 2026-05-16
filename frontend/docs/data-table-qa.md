# DataTable QA notes

QA-исследование multi-column поиска в [data-table.tsx](../src/components/ui/data-table.tsx) и связанной инфраструктуре. Документ — точка опоры для регрессионного тестирования любых будущих изменений в фильтре, picker'е, persistence-слоях и URL-контракте.

## Pages under test

| Path | Storage key | Filter mode | Searchable columns |
|---|---|---|---|
| `/flows` | `table_4_/flows` | controlled (`?q=`) | id, title, status, provider, terminals |
| `/knowledges` | `table_4_/knowledges` | controlled | docType, question, content |
| `/templates` | `table_4_/templates` | controlled | title, text |
| `/settings/providers` | `table_4_/settings/providers` | controlled | name, type |
| `/settings/api-tokens` | `table_4_/settings/api-tokens` | controlled | name, tokenId, status |
| `/settings/prompts` (agents) | `table_4_/settings/prompts` ⚠️ | uncontrolled | displayName, systemStatus, humanStatus |
| `/settings/prompts` (tools) | `table_4_/settings/prompts` ⚠️ | uncontrolled | displayName, status |

⚠️ Две таблицы делят один storage slot — см. Finding #3.

## Performance baseline (INP)

Измерения через Chrome DevTools MCP `performance_start_trace` / `performance_stop_trace`. Дев-сборка, без CPU/network throttling, 91 строка данных, печать ~6–10 символов подряд.

| Page | INP | Verdict |
|---|---|---|
| /flows | 41 ms | Good (<200ms) |
| /knowledges | 66 ms | Good |
| /templates | 30 ms | Good |
| /settings/providers | 77 ms | Good |
| /settings/api-tokens | 27 ms | Good |
| /settings/prompts | 78 ms | Good |

Историческая регрессия: до перехода на debounced input (commit `63d0598`) INP составлял 256–271 ms (input delay 237–253 ms) — печать ощущалась как лаги.

## Findings

### #1 — Stale `searchColumns` после расширения набора кандидатов

**Severity:** UX, low.

**Repro:** на странице с persisted `searchColumns: ['title', 'id', 'status']`, разработчик добавляет `meta.searchable: true` ещё двум колонкам. После reload picker показывает 5 кандидатов, но новые два **не отмечены**.

**Verdict:** by design (preservation of user intent). Если бы мы auto-добавляли — перечеркнули бы explicit choice пользователя. Discoverability could be improved (например, легкий visual hint у picker'а когда есть unused candidates), но это отдельная feature.

### #2 — Дублирующийся `id="data-table-search"` на /settings/prompts

**Severity:** bug, medium (a11y + DX).

**Repro:**
```js
document.querySelectorAll('#data-table-search').length // → 2
```

**Impact:** `getElementById` находит только первый input; screen readers могут запутаться; автотесты, ищущие по id, ломаются.

**Fix:** генерировать уникальный id через `React.useId()` (стабильный per-instance). Применить и к `name`-атрибуту.

### #3 — Обе таблицы /settings/prompts шерят один storage slot

**Severity:** bug, architectural.

**Repro:**
```js
// Sort first table by Agent Name (asc) — storage:
{ sorting: [{ id: 'displayName', desc: false }] }
// Sort second table by Tool Name (asc) — storage перезаписан:
{ sorting: [{ id: 'displayName', desc: false }] }
```

Sorting/columnVisibility/searchColumns обеих таблиц пишутся под единственный ключ `table_4_/settings/prompts` через `updateTableState`. Последний writer выигрывает; на reload одна таблица может получить чужое состояние, stale-rebase эффект частично перезатрёт его и опять последний writer победит. Race condition в зависимости от порядка mount'а.

**Сейчас не критично** только потому что у обеих таблиц колонки совпадают по id (`displayName` ↔ `displayName`, статусы разные но stale-rebase их вычищает). Расхождение появится сразу, как только id колонок разойдутся.

**Fix:** добавить prop `storageKey?: string` в `DataTable`, потребитель передаёт уникальный ключ per-table. Default — текущий `usePageStorageKeys().table` (zero-config для одиночных таблиц).

### #4 — Pagination button дропает `?q=` из URL

**Severity:** bug, **вне scope multi-column фичи** — относится к `usePagination`.

**Repro:**
```
/flows?q=bypass  →  click "last page" (>>)  →  /flows?page=6   ❌ q потерян
```

**Investigation:** [use-pagination.ts:86-101](../src/hooks/use-pagination.ts#L86-L101) использует functional updater `setSearchParams((previous) => new URLSearchParams(previous).set('page', ...))`. Functional form должен сохранять q, но q пропал. Возможно стейл-снапшот `previous` в react-router-flow.

Зафиксирован как backlog item — не блокирующий, но share-link с фильтром + pagination сломан.

### #5 — Нет лимита длины filter-query

**Severity:** edge case, low.

**Repro:** ввод 5000-символьной строки → URL `?q=AAA…AAA` length=5003. Browsers handle up to ~8000 byte URLs, серверы / Vite-proxy типично 2k–4k. Share-link становится unreliable.

**Fix opt:** ограничить max length input'а (например, 200 chars — намного больше любого реалистичного search query). Скорее всего overkill для текущего use case — оставить как low-priority backlog.

## Adversarial findings (negative results)

Не воспроизвели:

- **XSS** — React экранирует, `<script>alert(1)</script>` в input хранится как plain text, не выполняется. ✅
- **SQL injection style** (`' OR 1=1 --`) — клиентский filter без SQL, plain substring matching. ✅
- **Regex injection** (`.*+?[](){}|^$\`) — мы не используем regex, только `String.includes()`. ✅
- **Control chars** (`\n`, `\t`, `\0`) — `<input type=text>` стрипит `\n`, остальные сохраняются как plain string. ✅
- **Unicode / RTL / emoji** (`тест 🔥 العربية`) — URL-encoded, storage сохраняет UTF-8. ✅
- **Corrupted `searchColumns`** (`['nonexistent', 'title', '__proto__', 'ghost-id']`) — stale-rebase эффект вычищает невалидные id (включая `__proto__`), оставляет только валидные. Никаких prototype-pollution / crash'ей. ✅
- **Pagination `?page=999`** — clamping reconciles к last valid page. ✅
- **Back button** — typing keystrokes идут через `replace: true`, не засоряют history. ✅

## Regression test scenarios

Структура: **Setup → Steps → Expected**. Используется Chrome DevTools MCP, dev server на `https://localhost:8000`.

### S1 — Smoke на каждой странице

**Setup:** залогинен, `localStorage.clear()`, navigate to страница из таблицы выше.

**Steps:** на каждой из 6 страниц:
1. Открыть страницу.
2. Убедиться, что input `#data-table-search` отрендерен.
3. Убедиться, что кнопка `aria-label="Search in"` присутствует (только если на странице ≥2 searchable колонки).
4. Ввести подстроку, очевидно совпадающую только с одной колонкой.
5. Подождать 250 мс (debounce 150 мс + render).

**Expected:** строки сужаются до релевантных; `localStorage.getItem('table_4_<path>')` содержит `filter`; URL содержит `?q=<value>` (для controlled-mode страниц).

### S2 — Performance trace per page

**Setup:** dev server без CPU throttling.

**Steps:**
1. `mcp__chrome-devtools__performance_start_trace({ reload: false, autoStop: false })`.
2. Click input, `type_text('bypass')` (или другой 6–10-char string).
3. `mcp__chrome-devtools__performance_stop_trace()`.
4. Parse `INP` из summary.

**Expected:** INP ≤ 100 ms на всех 6 страницах. Регрессия — если INP > 200 ms (Good→Needs improvement boundary).

### S3 — Restore from localStorage

**Setup:** На странице ввести фильтр + узкий выбор в picker'е.

**Steps:**
1. Сделать narrow выбор (snять галочки).
2. Reload через `?ignoreCache=true`.
3. Проверить input value и picker checkboxes.

**Expected:** input восстановлен с предыдущим query, picker отражает сохранённый `searchColumns`.

### S4 — Stale-id rebase

**Setup:**
```js
localStorage.setItem('table_4_/flows', JSON.stringify({
  searchColumns: ['nonexistent-column', 'title', '__proto__', 'ghost-id']
}));
```

**Steps:** Reload страницы.

**Expected:** storage очищен до `{ searchColumns: ['title'] }` — единственного валидного id. Никаких crash'ей. Никакого prototype pollution.

### S5 — Adversarial input

**Setup:** Чистая страница `/flows`.

**Steps:** в input последовательно вводить (через `evaluate_script + setter + input event` для bypass userEvent latency):
1. `<script>alert(1)</script>`
2. `' OR 1=1 --`
3. `.*+?[](){}|^$\`
4. `тест 🔥 العربية`
5. `\n\t\0` mix
6. `'A'.repeat(5000)`

**Expected:** input принимает все значения как plain text, никаких alerts, никаких crashes, URL/localStorage сохраняют raw string без интерпретации. Для (6) — URL до ~5KB (без жёсткого max в текущей реализации).

### S6 — Pagination clamping

**Setup:** На странице с N rows, отфильтровать до 1–2 result'ов.

**Steps:**
1. Navigate to `?q=<filter>&page=999`.
2. Wait 500 мс для reconcile.

**Expected:** URL стал `?q=<filter>&page=<lastValidPage>`. Видна последняя страница.

### S7 — Back button skips typing history

**Setup:** Войти на страницу через прямой URL `?q=foo` (не через typing).

**Steps:**
1. В input ввести `bar` (поверх `foo`).
2. Press browser back.

**Expected:** Возврат на исходный URL `?q=foo` (не intermediate `?q=ba`, не `?q=b`). Typing keystrokes используют `replace: true`.

### S8 — Multi-tab / multi-instance

**Setup:** Открыть `/settings/prompts` — там 2 таблицы на одной странице.

**Steps:**
1. Sort first (Agents) table by Agent Name.
2. Sort second (Tools) table by Tool Name.
3. Reload.

**Expected (после Fix #3):** обе таблицы восстанавливают свой sort независимо. **До Fix #3** — sort одной из таблиц "побеждает" в storage, другая теряет сортировку.

### S9 — Cross-feature: filter + sort + pagination

**Setup:** `/flows`, clean state.

**Steps:**
1. Sort by ID desc.
2. Apply filter `bypass`.
3. Click "Page 3".
4. Click X button (clear filter).

**Expected:** sort сохраняется через все шаги; pagination сбрасывается на page 1 после filter change (`clearPageParamOnChange: true`); после clear — все строки видны, sort остался DESC.

### S10 — Picker semantics

**Setup:** `/flows`, clean state.

**Steps:**
1. Open picker, **снять все галочки**.
2. Type filter `bypass`.
3. Closed picker, re-open.

**Expected:** empty selection = sentinel "search everywhere" — поиск всё равно работает по всем кандидатам. Picker rendering: все checkboxes показаны checked (visualisation aligns with sentinel). storage содержит `searchColumns` отсутствующий (или массив с 0 элементов схлопывается в `updateTableState`).

### S11 — `id`/`name` uniqueness on multi-table page

**Setup:** `/settings/prompts`.

**Steps:**
```js
document.querySelectorAll('#data-table-search').length
document.querySelectorAll('input[name="search"]').length
```

**Expected (после Fix #2):** оба значения =2 если используется generated unique id per instance; или ровно 1 для `id`, 2 для `name` если применяем только id-fix. **Critical:** `getElementById('data-table-search')` не должен находить более одного элемента. **До Fix #2** оба равны 2 (баг).

## How to run

### Browser-based scenarios (S1–S11)

Используется Chrome DevTools MCP. Минимальная sequence:

```
mcp__chrome-devtools__navigate_page({ type: "url", url: "https://localhost:8000/flows" })
mcp__chrome-devtools__evaluate_script({ function: "() => { ... }" })   # для programmatic actions / assertions
mcp__chrome-devtools__performance_start_trace / type_text / performance_stop_trace
```

Используйте `evaluate_script` для:
- bulk programmatic actions (clear input, inject localStorage, sample state),
- избежания больших snapshot dumps.

Используйте `take_snapshot` только когда нужен UID элемента для `click` или другого pointer-event.

### Unit-level scenarios

Vitest покрывает picker rendering, filter behaviour, persistence, stale-rebase (см. [data-table.test.tsx](../src/components/ui/data-table.test.tsx)). Для debounce-sensitive проверок использовать `findBy*` / `waitFor`.

## Backlog

- [ ] **Fix #2** — unique `id`/`name` через `useId()`.
- [ ] **Fix #3** — prop `storageKey?: string` в DataTable, передать уникальные ключи в settings-prompts.
- [ ] **#4** — расследовать pagination loss of `?q=` (вне scope multi-column).
- [ ] **#5** — рассмотреть max length для input (low priority).
- [ ] **#1** — discoverability hint когда есть unused searchable candidates (low priority).
- [ ] Добавить unit-тест на сценарии S4 (stale-rebase) и S8 (multi-table isolation) когда Fix #3 будет применён.
- [ ] Когда появится Playwright/Cypress инфраструктура — портировать S2 (performance) и S6 (pagination) в e2e suite.
