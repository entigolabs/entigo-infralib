gcloud container clusters get-credentials pri-infra-gke --region europe-north1 --project entigo-infralib2
if [ $? -eq 0 ]
then
	echo "---"
	echo "Logged into Google PRI"
	echo "https://argocd-pri.pri-net-dns.gcp.infralib.entigo.io/"
        kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl config use-context gke_entigo-infralib2_europe-north1_pri-infra-gke"
fi
gcloud container clusters get-credentials biz-infra-gke --region europe-north1 --project entigo-infralib2
if [ $? -eq 0 ]
then
        echo "---"
        echo "Logged into Google BIZ"
        echo "https://argocd-biz.biz-net-dns.gcp.infralib.entigo.io/"
	kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl config use-context gke_entigo-infralib2_europe-north1_biz-infra-gke"
fi
aws eks update-kubeconfig --region eu-north-1 --name pri-infra-eks
if [ $? -eq 0 ]
then
        echo "---"
        echo "Logged into AWS PRI"
        echo "https://argocd-pri.pri-net-route53.infralib.entigo.io/"
	kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep -ve"ecr-updater" | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl config use-context arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks"
fi
aws eks update-kubeconfig --region eu-north-1 --name biz-infra-eks
if [ $? -eq 0 ]
then
        echo "---"
        echo "Logged into AWS BIZ"
        echo "https://argocd-biz.biz-net-route53.infralib.entigo.io/"
	kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep -ve"ecr-updater" | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl config use-context arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks"
fi
az aks get-credentials --resource-group pri-italynorth --name pri-infra-aks --overwrite-existing --only-show-errors && kubelogin convert-kubeconfig -l azurecli --context pri-infra-aks
if [ $? -eq 0 ]
then
        echo "---"
        echo "Logged into Azure PRI"
        echo "https://argocd-pri.pri-net-dns.azure.infralib.entigo.io/"
	kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl config use-context pri-infra-aks"
fi
az aks get-credentials --resource-group biz-italynorth --name biz-infra-aks --overwrite-existing --only-show-errors && kubelogin convert-kubeconfig -l azurecli --context biz-infra-aks
if [ $? -eq 0 ]
then
        echo "---"
        echo "Logged into Azure BIZ"
        echo "https://argocd-biz.biz-net-dns.azure.infralib.entigo.io/"
	kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl config use-context biz-infra-aks"
fi

