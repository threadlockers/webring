# threadlocked

a dead simple [webring](https://en.wikipedia.org/wiki/Webring) api server for threadlocked.

join our [discord server](https://discord.gg/gqyV3nyjwE) to meet cool engineers and hackers :p - please text @ni5arga or @datavorous_ on discord for approval or if you face any issues while joining. 

<div> 
<a href="https://discord.gg/gqyV3nyjwE">
  <img src="https://img.shields.io/badge/Join%20the%20Community-5865F2?style=for-the-badge&logo=discord&logoColor=white" />
</a>

</div>

## how to join the webring?

append your entry to [`webring.json`](./webring.json) and open a pr:

```json
{
  "name": "your-name",
  "gh": "your-github-username",
  "url": "https://your-site.com"
}
```

once your entry gets added to the webring, you can use `https://<name>.seggs.lol` as a redirect to your site.

## how to add prev / next / random to your site?

the base url for the server is `https://ring.seggs.lol`. to add previous/next/random buttons to your site, use the following html snippet:

```html
<!-- replace "your-name" with the exact name from webring.json -->
<a href="https://ring.seggs.lol/redirect?from=your-name&dir=prev">
  ← previous site
</a>

<a href="https://ring.seggs.lol/random">
  random site
</a>

<a href="https://ring.seggs.lol/redirect?from=your-name&dir=next">
  next site →
</a>
```
