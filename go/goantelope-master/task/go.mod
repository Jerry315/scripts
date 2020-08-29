module git.topvdn.com/web/goantelope/task

go 1.12

require (
	git.topvdn.com/web/goantelope/mongo v0.0.0-00010101000000-000000000000
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/google/uuid v1.1.1
	github.com/robfig/cron v0.0.0-20180505203441-b41be1df6967
	github.com/stretchr/testify v1.3.0
)

replace git.topvdn.com/web/goantelope/mongo => ../mongo
