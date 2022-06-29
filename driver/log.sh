#!/bin/bash
export TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
export APISERVER=https://api.fm-tok.fm-openshift.com:6443
export NAME=$(curl -sSk -H "Authorization: Bearer $TOKEN" $APISERVER/api/v1/namespaces/default/pods?labelSelector=app%3D$LABEL | jq .items[0].metadata.name | sed "s/\"//g")
echo $NAME
curl -sSk -H "Authorization: Bearer $TOKEN" $APISERVER/api/v1/namespaces/default/pods/$NAME/log?container=benchmark
