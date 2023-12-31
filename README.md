# Backend Technical Test at Scalingo

## Architecture
### Layout
There are 3 components:
* the updater queries the Github APIs and stores a flat result into Redis
* a Redis cache, used to store 100 latest repos, also caches any request coming to the front
* a front API that reads the Redis

This approach favors speed, and source of truth.
* A relational database was not deemed necessary for the requirements of this exercise, and would be slower.
* Also, the idea is to have a cached snapshot of the state of the latest 100 Github repos, not to create a duplicate of the Github DB. 
* Building a proxy that would query Github at each incoming request would not scale as easily.

### Querying the Github API
I found no Github API that gives the most 100 recently created repositories. There is:
* a /repositories endpoint that can give projects sorted by id, but only has 2 controls over pagination
    * a since parameter to get results since a certain project id
    * a Link header that has first page and next page links (not last)
* a /search/repositories endpoint that can give the repositories created above a certain date
    * this one does have a "last" link in the Link header
    * but, does not guarantee any sorted order, so the last page is not more relevant than another

The endpoint that can guarantee order is /repositories, so that one must be used. But, several pages of "next" links must be followed to get to the last page. A 'since' param must be found that is not too old so that we minimize the number of requests made.

So, this solution:
- calls /search/repositories?created>=[A RECENT TIME] to get the ID of a recently created repository
- then recursively call pages of /repositories?since=ID until the last page is found

Once the 100 latest repos are found, their language URLs are called concurrently.

At startup, and periodically, the updater caches a response, for the front's /repos endpoint
The front can filter by language, and will cache those results as well (like /repos?l=Python)

### Scaling
The images were created so they can be easily reused for a clustered deployment.
- The front and redis will scale well on Kubernetes with no change to the images
- The updater can become a Kubernetes Job with no change to the image, just set 
```
SBT_LOCAL_SCHEDULE=false
```

### Code
Coding to interfaces
Efficient algorithms (see parseLinks() for example)

## Execution

Add a valid Github API token in cmd/updater/local.env, like:
```
SBT_GITHUB_TOKEN="github_pat_etc..."
```
then:
```
docker compose up
```
Application will be then running on port `5000`

## Test

```
$ curl localhost:5000/repos
$ curl localhost:5000/repos?l=Python
```
