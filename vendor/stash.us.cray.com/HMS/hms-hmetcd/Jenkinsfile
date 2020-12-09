@Library('dst-shared@master') _

dockerBuildPipeline {
        repository = "cray"
        imagePrefix = "hms"
        app = "hmetcd"
        name = "hms-hmetcd"
        description = "Cray HMS hmetcd code."
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "internal"
}
