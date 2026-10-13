package report

// Suggestions are printed as text and never executed. None of them touches
// volumes: a volume holds data that no other Docker object reproduces.
const (
	SuggestBuilderPruneOld    = "docker builder prune --filter until=168h   # unused build cache older than 7 days"
	SuggestBuilderPruneAll    = "docker builder prune -a                    # all build cache not in use (next build is slower)"
	SuggestImagePruneDangling = "docker image prune                         # untagged (<none>) images only"
	SuggestImagePruneAll      = "docker image prune -a                      # also tagged images no container uses; they are re-pulled on demand"
	SuggestLogOptsDaemon      = `set "log-opts" in /etc/docker/daemon.json: {"log-driver":"json-file","log-opts":{"max-size":"10m","max-file":"3"}}   # new containers only`
	SuggestLogOptsContainer   = "recreate the container with --log-opt max-size=10m --log-opt max-file=3   # or --log-driver local"
)

// SuggestVolumeRm formats the single-volume removal hint. There is no
// bulk form on purpose.
func SuggestVolumeRm(name string) string {
	return "docker volume rm " + name + "   # only if you are sure nothing needs its data"
}
