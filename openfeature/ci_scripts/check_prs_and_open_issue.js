const {Octokit} = require('octokit');

const issueTitle = '👀 Open pull request on open feature contrib repositories'
const goffOwner = 'thomaspoignant'
const goffRepo = 'go-feature-flag'
const goffSlug = `${goffOwner}/${goffRepo}`
const issueHeader = '# Open Feature Provider pull requests 👀\n\nThis issues is the result of an automated process to look if there is anything related to GO Feature Flag open in the different Open Feature Contrib repos.\n\n🙏 Please close this issue when everything is merged.\n\n## Pull requests'
const commentHeader = '# Updated list of pull requests to 👀'
const repos = [
    {slug: 'open-feature/go-sdk-contrib', prefix: 'providers/go-feature-flag' },
    {slug: 'open-feature/java-sdk-contrib', prefix: 'providers/go-feature-flag' },
    {slug: 'open-feature/dotnet-sdk-contrib', prefix: 'src/OpenFeature.Contrib.Providers.GOFeatureFlag' },
    {slug: 'open-feature/js-sdk-contrib', prefix: 'libs/providers/go-feature-flag' },
]

const octokit = new Octokit({
    auth: process.env.GITHUB_TOKEN

});

async function fetchGOFFPullRequests(repoSlug, prefix){
    const goffPR = []
    const pulls = await octokit.request(`GET /repos/${repoSlug}/pulls?state=open`, {
        per_page: 100
    });
    const promises = pulls.data.map(pr => octokit.request(`GET /repos/${repoSlug}/pulls/${pr.number}/files`))
    const prs = await Promise.all(promises)

    return pulls.data.filter((item, index) => {
        const {data} = prs[index]
        return data.find(({filename})=> filename.startsWith(prefix))
    })
}

async function findAllPR(repos){
    let allPR= {}
    for (let index = 0; index < repos.length; index++) {
        const {slug, prefix} = repos[index]
        const open_prs = await fetchGOFFPullRequests(slug, prefix)
        const newObj = {
            [slug]: open_prs
        }
        allPR = { ...allPR, ...newObj}
    }
    return allPR
}


async function displayOpenPR(repos){
    const openPRs = await findAllPR(repos)
    const hasOpenPR = Object.keys(openPRs).length > 0
    let prListDisplay = ''
    if (!hasOpenPR) {
        return
    }

    Object.keys(openPRs).forEach((key) => {
        if(openPRs[key].length > 0){
            prListDisplay += `\n### ${key}\n\n`
            openPRs[key].forEach(({html_url})=> prListDisplay +=`- ${html_url}\n`)
        }
    })
    return prListDisplay
}

async function main(repos){
    const issues = await octokit.request(`GET /repos/${goffSlug}/issues?state=open`, {
        per_page: 1000
    });
    const notifIssue = issues.data.find(({title}) => title === issueTitle)
    const waitingPR = await displayOpenPR(repos)
    if (waitingPR !== ""){
        if (notifIssue === undefined){
            // Create new Issue
            await octokit.request(`POST /repos/${goffSlug}/issues`, {
                title: issueTitle,
                body: `${issueHeader} ${waitingPR}`,
                assignees: [
                    goffOwner
                ],
                labels: [
                    'dependencies'
                ]
            })
            console.log('issue created')
        } else {
            // Add comment in the issue
            await octokit.request(`POST /repos/${goffSlug}/issues/${notifIssue.number}/comments`, {
                body: `${commentHeader} ${waitingPR}`
            })
            console.log('comment added')
        }
    }

}

main(repos).then(console.log("success")).catch(err => {
    console.log(err)
    process.exit(1)
})