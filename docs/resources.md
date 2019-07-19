# Resources

## Kubernetes on WSL

On Windows command line run `kubectl config view` to find the values of [windows-user-name],[minikube-ip],[port]:

```sh
mkdir ~/.kube && cp /mnt/c/Users/[windows-user-name]/.kube/config ~/.kube
```

If you are using minikube you need to set bellow settings 
```sh
# allow kubectl to trust the certificate authority of minikube
kubectl config set-cluster minikube \
    --server=https://[minikube-ip]:[port] \
    --certificate-authority=/mnt/c/Users/[windows-user-name]/.minikube/ca.crt

# configure the client certificate to use when talking to minikube
kubectl config set-credentials minikube \
    --client-certificate=/mnt/c/Users/[windows-user-name]/.minikube/client.crt \
    --client-key=/mnt/c/Users/[windows-user-name]/.minikube/client.key

# create the context minikube with cluster and user info created above
kubectl config set-context minikube --cluster=minikube --user=minikub
```

More info:

- https://devkimchi.com/2018/06/05/running-kubernetes-on-wsl/
- https://www.jamessturtevant.com/posts/Running-Kubernetes-Minikube-on-Windows-10-with-WSL/

## Build pipelines

- [Create a pipeline and add a status badge to Github](https://docs.microsoft.com/en-us/azure/devops/pipelines/create-first-pipeline?view=azure-devops&tabs=tfs-2018-2)
- [Customize status badge with shields.io](https://shields.io/)