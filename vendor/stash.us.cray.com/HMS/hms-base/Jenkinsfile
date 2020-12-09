@Library('dst-shared@master') _

dockerBuildPipeline {
        repository = "cray"
        imagePrefix = "hms"
        app = "base"
        name = "hms-base"
        description = "Cray HMS base code."
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "internal"
}
