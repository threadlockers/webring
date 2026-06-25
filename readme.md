# threadlocked

a dead simple [webring](https://en.wikipedia.org/wiki/Webring) server for threadlocked. join our [discord server](https://discord.gg/gqyV3nyjwE) to meet cool engineers and hackers :p  

<div> 
<a href="https://discord.gg/gqyV3nyjwE">
  <img src="https://img.shields.io/badge/Join%20the%20Community-5865F2?style=for-the-badge&logo=discord&logoColor=white" />
</a>

</div>

## how to join the webring?

pull requests are restricted to users with write access only - please ping the admins in the [#webring](https://discord.com/channels/1463893045731921934/1510847098738704494) channel of the discord server to get yourself added into the webring. 

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

## badges

credits to [@nithitsuki](https://github.com/nithitsuki) for making these 88x31 badges

<img width="88" height="31" alt="image" src="https://github.com/user-attachments/assets/86fc434f-8b29-4628-b0a3-33cdf4d487c0" />

<img width="88" height="31" alt="image" src="https://github.com/user-attachments/assets/f4bd0e19-e15e-46ce-9505-2a373f579a73" />




