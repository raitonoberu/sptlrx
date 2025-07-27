# LRCLib Integration - V2 Architecture Preview

Cette branche démontre une architecture modulaire qui anticipe les besoins de sptlrx V2, tout en intégrant LRCLib comme nouvelle source de paroles.

## 🎯 Objectifs

1. **Intégrer LRCLib** - API gratuite et open source pour les paroles
2. **Architecture modulaire** - Préparer la transition vers V2
3. **Multi-sources** - Système de fallback intelligent
4. **Configuration flexible** - Structure TOML pour V2

## 🏗️ Architecture

### Structure Modulaire

```
services/
├── lrclib/           # Client LRCLib API
├── sources/          # Gestionnaire multi-sources
├── spotify/          # Client Spotify existant
└── examples/         # Exemples d'utilisation V2
```

### Interfaces Clés

#### `lyrics.Provider` (existante)

```go
type Provider interface {
    Lyrics(id, query string) ([]Line, error)
}
```

#### `sources.Manager` (nouvelle)

- Gestion multi-sources avec priorités
- Fallback automatique en cas d'échec
- Timeouts configurables par source
- Activation/désactivation dynamique

## 🔌 LRCLib API

### Avantages

- ✅ **Gratuite** - Pas de clé API requise
- ✅ **Open Source** - Communauté active
- ✅ **Paroles synchronisées** - Format LRC natif
- ✅ **Recherche robuste** - Par métadonnées exactes
- ✅ **Pas de rate limiting** - Usage libre

### Endpoints Utilisés

- `GET /api/get` - Recherche par métadonnées exactes
- `GET /api/get-cached` - Recherche dans cache seulement
- `GET /api/search` - Recherche par mots-clés

## 🎵 Sources Multiples - Vision V2

### Priorités Intelligentes

```
1. Spotify (Critique) - Si disponible, le plus précis
2. LRCLib (Haute) - Gratuit, fiable, communautaire
3. Genius (Moyenne) - Grandes collections
4. MusixMatch (Moyenne) - Commercial mais complet
5. AZLyrics (Basse) - Scraping de secours
```

### Logique de Fallback

1. **Essai par priorité** - Sources les plus fiables d'abord
2. **Timeouts individuels** - Éviter les blocages
3. **Retry intelligent** - Gestion des erreurs temporaires
4. **Cache intelligent** - Éviter les requêtes inutiles

## ⚙️ Configuration V2 (TOML)

```toml
[lyrics_sources]
enabled = ["lrclib", "spotify", "genius"]
priority = { spotify = 100, lrclib = 80, genius = 60 }
timeout = { spotify = "5s", lrclib = "10s", genius = "8s" }

[lyrics_sources.lrclib]
enabled = true
use_cached_only = false
custom_user_agent = "sptlrx/2.0.0"

[players]
default = "mpris"
priority = ["spotify", "mpris", "mpd"]
```

## 🚀 Migration Path V1 → V2

### Phase 1 (Actuelle)

- [x] Interface `lyrics.Provider` stable
- [x] Architecture modulaire services/
- [x] LRCLib comme provider alternatif

### Phase 2 (V2 Prep)

- [ ] Manager multi-sources intégré
- [ ] Configuration TOML
- [ ] Tests et benchmarks

### Phase 3 (V2)

- [ ] Remplacement lyricsapi → LRCLib
- [ ] Defaults intelligents (LRCLib + MPRIS)
- [ ] Architecture complètement modulaire

## 🧪 Tests et Développement

### Test LRCLib

```bash
# Test avec une chanson connue
curl "https://lrclib.net/api/get?artist_name=Borislav%20Slavov&track_name=I%20Want%20to%20Live&album_name=Baldur%27s%20Gate%203&duration=233"
```

### Build et Test

```bash
go build -o sptlrx .
./sptlrx --help
```

## 📈 Métriques et Monitoring

La nouvelle architecture permet :

- **Latency tracking** par source
- **Success rate** par provider
- **Fallback statistics** - Quelles sources fonctionnent le mieux
- **User preferences** - Sources préférées par usage

## 🤝 Contribution

Cette architecture facilite les contributions :

1. **Nouveau provider** = Implémentation de `lyrics.Provider`
2. **Configuration** = Ajout section TOML
3. **Tests** = Interface standardisée
4. **Documentation** = Structure claire

## 🔮 Roadmap

### Court Terme

- [ ] Parser LRC complet
- [ ] Tests unitaires
- [ ] Integration avec player existant

### Moyen Terme

- [ ] Genius API integration
- [ ] Cache persistant
- [ ] Metrics dashboard

### Long Terme (V2)

- [ ] Interface web de gestion
- [ ] Plugins system
- [ ] Cloud sync preferences

---

**Cette architecture anticipe V2 tout en restant compatible V1** ✨
