## Sets docker registry to point to local minishift
eval $(minishift docker-env)
docker login -u $(oc whoami) -p $(oc whoami -t) $(minishift openshift registry)
export REGISTRY=$(minishift openshift registry)
export DOCKER_REPO=$(oc get project --show-all=false -o=custom-columns=HOST:metadata.name --no-headers=true)

## Seed files required for OpenShift secrets
echo "${GH_WEBHOOK_SECRET}" > config/hmac.token # <1>
echo "${GH_TOKEN}" > config/oauth.token # <2>
echo "${SENTRY_DSN}" > config/sentry.dsn # <3>

## Handy aliases
alias oc-console='xdg-open https://$(minishift ip):8443/console &>/dev/null'
alias uh="ultrahook github http://$(oc get route/hook --no-headers=true -o=custom-columns=HOST:spec.host)/hook"