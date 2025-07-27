# Guide de Contribution sptlrx 🎵

## 🎯 Opportunités Identifiées

### 1. **Sources de Paroles Alternatives** (Priorité Haute)

**Objectif** : Ajouter des APIs ouvertes pour réduire la dépendance à Spotify

**Implémentation suggérée** :

```
services/
├── genius/          # API Genius (gratuite)
├── musixmatch/      # API MusixMatch
├── lrclib/          # LRCLib (open source)
└── azlyrics/        # Scraping AZLyrics
```

**Points d'entrée** :

- Interface `lyrics.Provider` dans `lyrics/lyrics.go`
- Ajouter nouveaux services dans `services/`
- Configuration dans `config/`

### 2. **App Multiplateforme** (Issue #26)

**Options** :

- **Web App** : Interface browser avec WebSocket
- **Desktop** : Utiliser Wails ou Electron + Go backend
- **Mobile** : Flutter avec plugin Go

### 3. **Améliorations UX** (Issues actives)

- Support souris (#65)
- Indication "paroles introuvables" (#55)
- Multi-players simultanés (#63)
- Support macOS amélioré (#17)

## 🔧 Configuration de Développement

### Installation des Outils Go

```bash
# Tools pour le développement
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Structure du Projet

```
sptlrx/
├── cmd/             # CLI commands (cobra)
├── config/          # Configuration management
├── lyrics/          # Interface lyrics.Provider
├── player/          # Interface player.Player
├── services/        # Implémentations concrètes
├── ui/              # Terminal UI (bubbletea)
└── pool/            # Connection pooling
```

## 🚀 Plan de Contribution Suggéré

### Phase 1 : Nouvelle Source de Paroles (LRCLib)

1. **Recherche** : Étudier l'API LRCLib (open source, gratuite)
2. **Implémentation** : Créer `services/lrclib/lrclib.go`
3. **Configuration** : Ajouter option dans config.yaml
4. **Tests** : Écrire tests unitaires
5. **Documentation** : Mettre à jour README.md

### Phase 2 : Améliorations UX

1. **Indication "pas de paroles"** (Issue #55)
2. **Support multi-sources** (fallback automatique)
3. **Cache local** pour améliorer les performances

### Phase 3 : App Web (Issue #26)

1. **Backend WebSocket** : Serveur Go pour streaming
2. **Frontend** : Interface web responsive
3. **Intégration** : Commande `sptlrx web`

## 📋 Checklist de Contribution

### Avant de Commencer

- [ ] Fork du repository
- [ ] Créer une branche feature
- [ ] Lire les issues existantes
- [ ] Étudier le code existant

### Développement

- [ ] Suivre les conventions Go
- [ ] Respecter l'architecture modulaire
- [ ] Écrire des tests
- [ ] Documenter les APIs publiques
- [ ] Tester manuellement

### Soumission

- [ ] Commit messages clairs
- [ ] Rebase sur master
- [ ] Tests passing
- [ ] Documentation à jour
- [ ] Pull Request avec description

## 🛠️ Commandes Utiles

```bash
# Build et test
go build -o sptlrx .
go test ./...
go mod tidy

# Linting
golangci-lint run
staticcheck ./...

# Debug avec verbose
./sptlrx --verbose --player mpris

# Test pipe mode
./sptlrx pipe
```

## 📚 Ressources

- [Issues GitHub](https://github.com/raitonoberu/sptlrx/issues)
- [Documentation Go](https://golang.org/doc/)
- [Bubble Tea (UI)](https://github.com/charmbracelet/bubbletea)
- [Cobra (CLI)](https://github.com/spf13/cobra)

---

**Prêt à contribuer ?** Commencez par explorer une issue qui vous intéresse ! 🚀
