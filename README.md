# Argo CD Namespace Operator

The Argo CD namespace operator helps creating the namespace for an [Argo CD `Application` CR](https://github.com/argoproj/argo-cd/blob/master/manifests/crds/application-crd.yaml).

Annotations on the `Application` CR allow defining labels and annotations for the `Namespace` CR. Special support is included for the Rancher `field.cattle.io/projectId` label and annotation to define a namespace as part of a Rancher project.

## Examples

### Define a label on the `Namespace` CR
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: cert-manager
  annotations:
    argocd-namespace.topicus.nl/label: "certmanager.k8s.io/disable-validation: true"
  finalizers:
  - resources-finalizer.argocd.argoproj.io
spec:
  destination:
    namespace: cert-manager
    server: https://kubernetes.default.svc
  ...
```
This `Application` CR automatically created the following namespace:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: cert-manager
  labels:
    certmanager.k8s.io/disable-validation: "true"
```

### Combine with the Namespace Configuration Operator
The [Namespace Configuration Operator](https://github.com/redhat-cop/namespace-configuration-operator) uses a (label)selector to select namespaces to which to apply its configuration. These labels can be supplied with the Argo CD namespace operator from the `Application` CR.
See the [T-shirt sized quotas](https://github.com/redhat-cop/namespace-configuration-operator#t-shirt-sized-quotas) example of the Namespace Configuration Operator.
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-large-project
  annotations:
    argocd-namespace.topicus.nl/label: "size: large"
  finalizers:
  - resources-finalizer.argocd.argoproj.io
spec:
  destination:
    namespace: large-project
    server: https://kubernetes.default.svc
  ...
```

### Rancher integration
The `field.cattle.io/projectId` label and annotation binds a namespace to a Rancher project. The full Rancher project ID (including the cluster ID) can be supplied with the `cattle.topicus.nl/projectId` annotation.
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-project
  annotations:
    cattle.topicus.nl/projectId: c-xxxxx:p-yyyyy
  finalizers:
  - resources-finalizer.argocd.argoproj.io
spec:
  destination:
    namespace: rancher-project-namespace
    server: https://kubernetes.default.svc
  ...
```

This creates the required Rancher label and annotation as follows:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: rancher-project-namespace
  annotations:
    field.cattle.io/projectId: c-xxxxx:p-yyyyy
  labels:
    field.cattle.io/projectId: p-yyyyy
```

## Links
- [Operator SDK](https://github.com/operator-framework/operator-sdk)
- [Argo CD](https://argoproj.github.io/argo-cd/)
- [Rancher](https://rancher.com/)
- [Namespace Configuration Operator](https://github.com/redhat-cop/namespace-configuration-operator)
