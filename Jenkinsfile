@Library('dst-shared@master') _

dockerBuildPipeline {
        githubPushRepo = "Cray-HPE/hms-firmware-action"
  		githubPushBranches =  /(release\/.*|master)/
        repository = "cray"
        imagePrefix = "cray"
        app = "firmware-action"
        name = "hms-firmware-action"
        description = "Cray firmware action service"
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "csm"
}
