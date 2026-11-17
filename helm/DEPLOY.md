# Exemples de déploiement du Helm Chart

## 1. Installation en développement

```bash
helm install dev-release ./helm -f helm/values.dev.yaml
```

## 2. Installation en production

```bash
helm install prod-release ./helm -f helm/values.prod.yaml \
  --set secrets.github.token="ghp_xxxxxxxxxxxx" \
  --set secrets.s3.accessKeyId="AKIA..." \
  --set secrets.s3.secretAccessKey="wJalr..."
```

## 3. Avec un namespace

```bash
helm install my-release ./helm -n production --create-namespace
```

## 4. Helm values via fichier secrets

Créer un fichier `secrets.yaml`:

```yaml
secrets:
  github:
    token: "ghp_xxxxxxxxxxxx"
  s3:
    accessKeyId: "AKIAIOSFODNN7EXAMPLE"
    secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
```

Puis:

```bash
helm install my-release ./helm -f helm/values.prod.yaml -f secrets.yaml
```

**IMPORTANT**: Ne pas commiter `secrets.yaml` !

## 5. Utiliser External Secrets Operator

Pour éviter de stocker les secrets en plaintext:

```bash
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets-system \
  --create-namespace
```

Créer un fichier `external-secret.yaml`:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: github-secret-store
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "my-app"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: mozilla-extension-gate-secrets
spec:
  secretStoreRef:
    name: github-secret-store
    kind: SecretStore
  target:
    name: mozilla-extension-gate
    creationPolicy: Owner
  data:
    - secretKey: github-token
      remoteRef:
        key: github/token
    - secretKey: s3-access-key-id
      remoteRef:
        key: s3/access-key
    - secretKey: s3-secret-access-key
      remoteRef:
        key: s3/secret-key
```

Puis appliquer et déployer:

```bash
kubectl apply -f external-secret.yaml
helm install my-release ./helm -f helm/values.prod.yaml
```

## 6. Helm values via ArgoCD

Fichier `argocd-application.yaml`:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: mozilla-extension-gate
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/mozilla-extension-gate
    targetRevision: HEAD
    path: helm
    helm:
      releaseName: mozilla-extension-gate
      values: |
        config:
          server:
            publicUrl: https://api.example.com
          github:
            repository: myorg/myrepo
      valueFiles:
        - values.prod.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

## 7. Vérification du déploiement

```bash
# Lister les releases
helm list

# Vérifier l'état du déploiement
kubectl get deployments
kubectl get pods
kubectl get services

# Voir les logs
kubectl logs -f deployment/mozilla-extension-gate

# Accéder à l'application
kubectl port-forward service/mozilla-extension-gate 8080:80
# puis ouvrir http://localhost:8080
```

## 8. Mise à jour

```bash
helm upgrade my-release ./helm -f helm/values.prod.yaml
```

## 9. Rollback

```bash
helm rollback my-release
```

## 10. Supprimer

```bash
helm uninstall my-release
```
