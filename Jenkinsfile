@Library('dst-shared@release/shasta-1.4') _

dockerBuildPipeline {
        repository = "cray"
        imagePrefix = "cray"
        app = "firmware-action"
        name = "hms-firmware-action"
        description = "Cray firmware action service"
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "csm"
}
