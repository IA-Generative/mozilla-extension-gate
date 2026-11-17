# Mozilla Extension Gate Helm Chart

Helm chart pour déployer l'application mozilla-extension-gate sur Kubernetes.

## Installation

### Installation basique

```bash
helm install my-release ./helm
```

### Installation avec des valeurs personnalisées

```bash
helm install my-release ./helm \
  --set config.server.publicUrl="https://api.example.com" \
  --set config.github.repository="myorg/myrepo" \
  --set secrets.github.token="ghp_xxxxxxxxxxxx" \
  --set config.s3.bucket="my-bucket"
```

## Configuration

Toutes les configurations de l'application sont gérées via le fichier `values.yaml`.

### Configuration du serveur

```yaml
config:
  server:
    env: "production"              # Environnement (development/production)
    port: "8080"                   # Port d'écoute
    publicUrl: "http://api.local"  # URL publique
    basePath: "/"                  # Chemin de base
    version: "1.0"                 # Version de l'app
    instanceId: "gateway-instance" # Identifiant d'instance
    artifactBaseUrl: ""            # URL de base des artefacts
    proxyArtifactEnabled: false    # Activer le proxy d'artefacts
```

### Configuration GitHub

```yaml
config:
  github:
    repository: "org/repo"      # Repository GitHub
    syncInterval: "5m"          # Intervalle de sync
```

Les secrets doivent être fournis séparément:

```yaml
secrets:
  github:
    token: "ghp_xxxxxxxxxxxx"   # Token GitHub (base64 encodé automatiquement)
```

### Configuration S3

```yaml
config:
  s3:
    bucket: "my-bucket"          # Nom du bucket
    region: "us-east-1"           # Région AWS
    endpoint: "https://s3.aws..." # Endpoint S3

secrets:
  s3:
    accessKeyId: "AKIAIOSFODNN7EXAMPLE"
    secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
```

## Déploiement avec values override

### Via fichier values

Créer un fichier `my-values.yaml`:

```yaml
config:
  server:
    env: production
    publicUrl: https://api.example.com
  github:
    repository: myorg/myrepo
  s3:
    bucket: my-bucket
    region: eu-west-1
    endpoint: https://s3.example.com

secrets:
  github:
    token: ghp_xxxxxxxxxxxx
  s3:
    accessKeyId: AKIA...
    secretAccessKey: wJalr...
```

Puis déployer:

```bash
helm install my-release ./helm -f my-values.yaml
```

### Variables d'environnement additionnelles

Pour ajouter des variables d'environnement supplémentaires:

```yaml
env:
  - name: DEBUG
    value: "true"
  - name: LOG_LEVEL
    value: "debug"
```

## Mise à jour

```bash
helm upgrade my-release ./helm -f my-values.yaml
```

## Désinstallation

```bash
helm uninstall my-release
```

## Gestion des secrets

Les secrets (GitHub token, S3 credentials) sont gérés de manière flexible:

### Option 1: Plaintext dans values (développement uniquement)

```yaml
secrets:
  github:
    token:
      type: "value"
      value: "ghp_xxxxxxxxxxxx"
  s3:
    accessKeyId:
      type: "value"
      value: "AKIA..."
    secretAccessKey:
      type: "value"
      value: "wJalr..."
```

**⚠️ DANGER**: Ne jamais commiter en plaintext en production!

### Option 2: Référencer des Secrets Kubernetes existants (RECOMMANDÉ)

```yaml
secrets:
  github:
    token:
      type: "secretRef"
      secretRef:
        name: github-credentials
        key: token
  s3:
    accessKeyId:
      type: "secretRef"
      secretRef:
        name: s3-credentials
        key: access-key-id
    secretAccessKey:
      type: "secretRef"
      secretRef:
        name: s3-credentials
        key: secret-access-key
```

Créer les secrets Kubernetes avant le déploiement:

```bash
kubectl create secret generic github-credentials \
  --from-literal=token="ghp_xxxxxxxxxxxx"

kubectl create secret generic s3-credentials \
  --from-literal=access-key-id="AKIA..." \
  --from-literal=secret-access-key="wJalr..."
```

### Option 3: External Secrets Operator (AWS Secrets Manager, Vault, etc.)

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets -n external-secrets-system --create-namespace
```

Puis créer une SecretStore et ExternalSecret.

## Ressources

Les ressources par défaut sont configurées pour un environnement de développement. Pour la production:

```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

## Autoscaling

Pour activer l'autoscaling:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```

## Ingress

Pour exposer l'application via Ingress:

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: api-tls
      hosts:
        - api.example.com
```
