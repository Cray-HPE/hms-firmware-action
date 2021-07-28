@Library('dst-shared@master') _

dockerBuildPipeline {
        githubPushRepo = "Cray-HPE/hms-hmetcd"
        repository = "cray"
        imagePrefix = "hms"
        app = "hmetcd"
        name = "hms-hmetcd"
        description = "Cray HMS hmetcd code."
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "internal"
}
