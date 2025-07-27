# Instructions Copilot pour sptlrx

<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

## Contexte du projet

sptlrx est une application CLI Go qui affiche des paroles synchronisées dans le terminal.

### Architecture

- **Langage** : Go
- **Structure modulaire** :
  - `cmd/` : Commandes CLI
  - `config/` : Gestion de la configuration
  - `lyrics/` : Récupération et traitement des paroles
  - `player/` : Intégration avec différents players
  - `services/` : Services externes (APIs)
  - `ui/` : Interface utilisateur terminal
  - `pool/` : Gestion des connexions

### Bonnes pratiques

- Utiliser les conventions Go standard
- Respecter l'architecture modulaire existante
- Tester les nouvelles fonctionnalités
- Documenter les APIs publiques
- Suivre les patterns existants pour la gestion d'erreurs

### Fonctionnalités principales

- Support multi-players : Spotify, MPD, Mopidy, MPRIS, Browser
- Affichage terminal avec styles personnalisables
- Mode pipe pour intégration avec d'autres outils
- Configuration YAML flexible
- Support des fichiers LRC locaux

### Opportunités de contribution

- Nouvelles sources de paroles (APIs ouvertes)
- Amélioration du support multi-plateforme
- Interface web/GUI
- Support de nouveaux players
- Optimisations de performance
