# Plan — `komlist`, une CLI de gestion de tâches en Go

> Document de référence du plan, en attente de validation explicite avant implémentation.

## Contexte

Le dossier `/Users/montgomery/Documents/Poc/komlist/` est (était) vide. L'objectif est de créer **de zéro** un outil CLI en Go nommé `komlist` qui permet de :

- créer des tâches (`add`)
- les lister, avec filtre optionnel par statut (`list`)
- changer leur statut (`status`)
- les supprimer (`delete`)

Choix proposés (à valider) :

- **Framework CLI** : [Cobra](https://github.com/spf13/cobra) (sous-commandes, flags, aide auto).
- **Stockage** : fichier JSON local dans `~/.komlist/tasks.json`. Pas de base de données, lisible à la main.
- **Statuts** : `todo`, `in-progress`, `blocked`, `done`.
- **Commandes** : `add`, `list`, `status`, `delete`.

L'outil doit être installable via `go install` et utilisable comme `komlist <commande>`.

---

## Arborescence cible

```
komlist/
├── go.mod
├── go.sum
├── main.go                  # point d'entrée → cmd.Execute()
├── cmd/
│   ├── root.go              # commande racine Cobra
│   ├── add.go               # komlist add "titre"
│   ├── list.go              # komlist list [--status xxx]
│   ├── status.go            # komlist status <id> <nouveau-statut>
│   └── delete.go            # komlist delete <id>
├── internal/
│   ├── task/
│   │   └── task.go          # type Task + type Status + validation
│   └── store/
│       └── store.go         # Load / Save / NextID / chemin du fichier
└── README.md
```

Séparation `cmd/` (interface CLI) vs `internal/` (logique métier et persistance) — facilite les tests et empêche l'import externe de `internal/`.

---

## Modèle de données

### `internal/task/task.go`

```go
type Status string

const (
    StatusTodo       Status = "todo"
    StatusInProgress Status = "in-progress"
    StatusBlocked    Status = "blocked"
    StatusDone       Status = "done"
)

func (s Status) Valid() bool { /* whitelist */ }
func AllStatuses() []Status   { /* pour l'aide CLI et la validation */ }

type Task struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Status    Status    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### `internal/store/store.go`

Fichier JSON unique contenant la liste des tâches et le compteur d'ID.

```go
type Store struct {
    Tasks  []task.Task `json:"tasks"`
    NextID int         `json:"next_id"`
}

func Path() (string, error)            // ~/.komlist/tasks.json (via os.UserHomeDir)
func Load() (*Store, error)            // crée le fichier vide si absent
func (s *Store) Save() error           // écriture atomique : tmp + os.Rename
func (s *Store) Add(title string) task.Task
func (s *Store) Find(id int) (*task.Task, error)
func (s *Store) SetStatus(id int, st task.Status) error
func (s *Store) Delete(id int) error
```

Points clés :

- Le répertoire `~/.komlist/` est créé à la volée (`os.MkdirAll`, perms `0o755`).
- Écriture atomique pour ne jamais corrompre le fichier en cas d'interruption.
- Permissions fichier `0o644`.

---

## Commandes Cobra

### `cmd/root.go`

```go
var rootCmd = &cobra.Command{
    Use:   "komlist",
    Short: "Gestion de tâches en ligne de commande",
}

func Execute() { /* rootCmd.Execute() + os.Exit en cas d'erreur */ }
```

Chaque sous-commande s'enregistre dans son `init()` via `rootCmd.AddCommand(...)`.

### `add`

```
komlist add "Acheter du pain"
```

- Args : exactement 1 (le titre). `cobra.ExactArgs(1)`.
- Statut par défaut : `todo`.
- Affiche : `Créée: #3 [todo] Acheter du pain`.

### `list`

```
komlist list
komlist list --status in-progress
```

- Flag `--status` / `-s` optionnel, validé contre `AllStatuses()`.
- Affichage tabulaire aligné via `text/tabwriter` (stdlib).
- Colonnes : `ID`, `STATUT`, `TITRE`, `MIS À JOUR` (format relatif simple ou RFC3339).
- Si liste vide : message `Aucune tâche.`.

### `status`

```
komlist status 3 in-progress
```

- Args : exactement 2 — id (int) et statut. `cobra.ExactArgs(2)`.
- Erreur claire si id introuvable ou statut invalide (avec la liste des statuts valides).
- Met à jour `UpdatedAt`.

### `delete`

```
komlist delete 3
```

- Args : exactement 1 (id). `cobra.ExactArgs(1)`.
- Erreur claire si id introuvable.
- Pas de confirmation interactive (CLI scriptable). Possibilité d'ajouter un flag `--yes` plus tard si besoin.

---

## Gestion des erreurs

- Toutes les commandes retournent une `error` via la signature `RunE` de Cobra (et non `Run`). Cobra affiche l'erreur et exit avec un code ≠ 0.
- Erreurs typées simples avec `fmt.Errorf("tâche %d introuvable", id)` — pas besoin de hiérarchie d'erreurs pour ce POC.

---

## Étapes d'implémentation

1. **Init module** : `go mod init github.com/montgomery/komlist` (ou nom choisi).
2. **Dépendance** : `go get github.com/spf13/cobra@latest`.
3. **`internal/task/task.go`** : type `Task`, type `Status`, helpers de validation.
4. **`internal/store/store.go`** : Load/Save atomique, Add/Find/SetStatus/Delete.
5. **`cmd/root.go`** + `main.go` : squelette Cobra exécutable.
6. **`cmd/add.go`**, **`cmd/list.go`**, **`cmd/status.go`**, **`cmd/delete.go`** : une commande par fichier.
7. **`README.md`** : section installation (`go install ./...`) + exemples des 4 commandes.

---

## Vérification end-to-end

Après build (`go build -o komlist .`) :

```bash
./komlist add "Écrire le plan"
./komlist add "Implémenter la CLI"
./komlist list
./komlist status 2 in-progress
./komlist list --status in-progress
./komlist status 1 done
./komlist delete 1
./komlist list
cat ~/.komlist/tasks.json     # vérifier le JSON sur disque
```

Cas d'erreur à valider manuellement :

- `./komlist status 999 done` → message "tâche 999 introuvable", exit ≠ 0.
- `./komlist status 1 foo` → message listant les statuts valides, exit ≠ 0.
- `./komlist list --status foo` → même validation que ci-dessus.

Si tout passe, installation système :

```bash
go install .
komlist list
```

---

## Hors scope (volontairement)

- Tests automatisés (peuvent être ajoutés ensuite ; le scope demandé est l'outil lui-même).
- Édition du titre (`edit`), tags, dates d'échéance, priorités.
- TUI interactive (Bubble Tea).
- Multi-utilisateurs / sync distante.

---

## État actuel du dossier (déjà fait avant pause)

- `go mod init github.com/montgomery/komlist` exécuté → `go.mod` présent.
- `go get github.com/spf13/cobra@latest` → `go.sum` présent.
- Aucun fichier `.go` écrit.

Si tu veux repartir d'une feuille blanche, supprime `go.mod` et `go.sum` ; sinon ils sont prêts à servir tels quels.
