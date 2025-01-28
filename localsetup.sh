gcloud container clusters get-credentials pri-infra-gke --region europe-north1
if [ $? -eq 0 ]
then
	echo "---"
	echo "Logged into Google PRI"
	echo "https://argocd.pri-net-dns.gcp.infralib.entigo.io/"
        kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl --context='gke_entigo-infralib2_europe-north1_pri-infra-gke' -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get applications"
fi
gcloud container clusters get-credentials biz-infra-gke --region europe-north1
if [ $? -eq 0 ]
then
        echo "---"
        echo "Logged into Google BIZ"
        echo "https://argocd.biz-net-dns.gcp.infralib.entigo.io/"
	kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl --context='gke_entigo-infralib2_europe-north1_biz-infra-gke' -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get applications"
fi
aws eks update-kubeconfig --region eu-north-1 --name pri-infra-eks
if [ $? -eq 0 ]
then
        echo "---"
        echo "Logged into AWS PRI"
        echo "https://argocd.pri-net-route53.infralib.entigo.io/"
	kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl --context='arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks' -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get applications"
fi
aws eks update-kubeconfig --region eu-north-1 --name biz-infra-eks
if [ $? -eq 0 ]
then
        echo "---"
        echo "Logged into AWS BIZ"
        echo "https://argocd.biz-net-route53.infralib.entigo.io/"
	kubectl -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
	echo ""
	echo "kubectl --context='arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks' -n $(kubectl get namespaces -o custom-columns=':metadata.name' | grep ^argocd) get applications"
fi


