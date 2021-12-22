👋  Hey! Here is the image we built for you ([Artifactory Link](https://artifactory.algol60.net/ui/repos/tree/General/csm-docker%2F{{ .stableRepo }}%2F{{ .imageName }}%2F{{ .imageTag }})):

```bash
{{ .imageDownloadLink }}
```

Use podman or docker to pull it down and inspect locally:

```bash
podman pull {{ .imageDownloadLink }}
```

Or, use this script to pull the image from the build server to a dev system:

<details>
<summary>Dev System Pull Script</summary>
<br />

```
#!/usr/bin/env bash
export REMOTE_IMAGE={{ .fullImageWithShaTag }}
export LOCAL_IMAGE={{ .imageWithShaTag }}
export SLES_SP=SP2
zypper addrepo https://slemaster.us.cray.com/SUSE/Products/SLE-Module-Server-Applications/15-${SLES_SP}/x86_64/product {{ .zypperRepoName }}
zypper refresh
zypper in -y --repo {{ .zypperRepoName }} skopeo
skopeo copy --dest-tls-verify=false docker://${REMOTE_IMAGE} docker://registry.local/cray/${LOCAL_IMAGE}
zypper rr {{ .zypperRepoName }}
```
</details>

<details>
<summary>Snyk Report</summary>
<br />

_Coming soon_

</details>

<details>
<summary>Software Bill of Materials</summary>
<br />

_Coming soon_

</details>

*Note*: this SHA is the merge of {{ .PRHeadSha }} and the PR base branch. Good luck and make rocket go now! 🌮 🚀