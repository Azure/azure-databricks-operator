# Resources

## Kubernetes on WSL

On windows command line run `kubectl config view` to find the values of [windows-user-name],[minikubeip],[port]

```shell
mkdir ~/.kube \
&& cp /mnt/c/Users/[windows-user-name]/.kube/config ~/.kube
```

if you are using minikube you need to set bellow settings 
```shell
kubectl config set-cluster minikube --server=https://<minikubeip>:<port> --certificate-authority=/mnt/c/Users/<windows-user-name>/.minikube/ca.crt
kubectl config set-credentials minikube --client-certificate=/mnt/c/Users/<windows-user-name>/.minikube/client.crt --client-key=/mnt/c/Users/<windows-user-name>/.minikube/client.key
kubectl config set-context minikube --cluster=minikube --user=minikub

```

More info:

1. https://devkimchi.com/2018/06/05/running-kubernetes-on-wsl/
2. https://www.jamessturtevant.com/posts/Running-Kubernetes-Minikube-on-Windows-10-with-WSL/

## Build pipelines
1. [Create a pipeline and add a status badge to Github](https://docs.microsoft.com/en-us/azure/devops/pipelines/create-first-pipeline?view=azure-devops&tabs=tfs-2018-2)
2. [Customize status badge with shields.io](https://shields.io/)