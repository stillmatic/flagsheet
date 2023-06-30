# FeatureSheet

The easiest way to do feature flagging - just use Google sheets!

It's as easy as creating a new spreadsheet like this:

| Key              | Layer | Value | Weight (sum to 1000 per layer) |
|------------------|-------|-------|-------------------------------|
| my_key           | a     | foo   | 250                           |
| my_key           | a     | bar   | 750                           |
| my_other_key     | b     | foo   | 400                           |
| my_other_key     | b     | bar   | 100                           |
| overlapping_key  | b     | baz   | 10                            |
| overlapping_key  | b     | car   | 10                            |
| overlapping_key  | b     | dag   | 480                           |

(we actually don't use the key at all, but it's useful for humans to read)

The library can be used as an in-memory cache like this:

```go
fv, ok := fs.Evaluate("my_key", "user123")
if !ok {
    // error handling
}
switch fv {
case "foo":
    // do something
case "bar":
    // do something else
}
```

Or as a service, which you can connect to from any language via the excellent [Connect](https://connect.build/) platform, including via just CURL / REST. 


The library will automatically refresh the cache on a cadence of your choosing. 


### Features

- Lowest common denominator -- everybody can use Google Sheets
- Easy evaluation API
    - We use [Connect](https://connect.build/) for high performance, language-agnostic serving
    - Generating clients is very easy in different languages
- Approximately free to use
    - API is free, very very low memory and CPU usage, no storage
- Reasonable defaults and error handling
    - If weights sum over 1000, we will throw an error
    - If weights sum under 1000, we will return empty string for default values
- Built-in and free audit logging
    - Just check the Google sheets revision history
- Bulit-in and free RBAC
    - Just use Google Sheets permissions
- MIT license

Much of the concurrency logic is borrowed from [go-cache](https://github.com/patrickmn/go-cache/tree/master). We optimize it slightly as we refresh the entire cache at once, rather than on a per-key basis.


An extremely simple benchmark - 

```
cpu: AMD Ryzen 9 7950X 16-Core Processor            
BenchmarkEvaluate-32    	 4331403	       283.5 ns/op	      96 B/op	       3 allocs/op
```

Via the loadtest GRPC client at 100 qps / 10 seconds

```
2023/06/30 15:01:53 total: 206.777206ms, count: 1000
2023/06/30 15:01:53 avg: 206.777µs, p90: 295.25µs, p99: 575.208µs
```

It's literally an in-memory cache, I don't think it can get much faster (or simpler).

### Caveats

I would be very very careful using this for serious experimentation. Please do not lecture me, I studied statistics and was a professional data scientist, I am fully aware of the experimentation pitfalls. You will likely make mistakes relying on this library -- but you will also likely make mistakes using ANY experimentation platform.

When I was still a Professional Data Scientist, I simulated [the math here](https://twitter.com/hingeloss/status/1189286349901324288). Basically, as long as you believe that your changes aren't terrible, you should heavily weight towards shipping (p=0.50 vs 0.05 criterion). In general, people are far, far too conservative with experiments. Of course, if you don't do the math right, who knows if the p-value is remotely valid, I guess that's fair. 

Anyways, I feel comfortable using this in prod :D - there are enough other places that we'll mess up that it's probably not this.

# Usage 

Before getting started, you'll need to create a service account in the workspace that you're using. Download the JSON file and save it as `client_secret.json` (it will be named something like `adjective-noun-random-string.json`) in the root of your project. Share the spreadsheet with the service account email address.

Then, instantiate a Google sheets client:

```go
data, err := os.ReadFile("client_secret.json")
assert.NoError(t, err)

conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
assert.NoError(t, err)

client := conf.Client(context.TODO())
service := spreadsheet.NewServiceWithClient(client)
```

Instantiate a FeatureSheet client:

```go
spreadsheetID := "15_oV5NcvYK7wK3VVD5ol6KVkWHzPLFl22c1QyLYplpU"
fs, err := featuresheet.NewFeatureSheet(service, spreadsheetID, 1*time.Second)
assert.NoError(t, err)
assert.NotNil(t, spreadsheet)
fv, ok := fs.Get("custom_backend")
assert.True(t, ok)
assert.NotEmpty(t, fv)
```

You can view an [example sheet](https://docs.google.com/spreadsheets/d/15_oV5NcvYK7wK3VVD5ol6KVkWHzPLFl22c1QyLYplpU/edit#gid=0).

# Notes

Look, it uses Google Sheets. There a million bad things from there, so you know, be aware.
- Note that the Google Sheets API has a [rate limit](https://developers.google.com/docs/api/limits) that you must respect. 
- Note also that refreshes are somewhat slow - the API is slow and sheets parsing is unoptimized. 
- In practice I don't think this should matter much, but suggest a 10 second refresh interval. This should be very safe in terms of rate limit, and also mean you don't have to think too much about the race conditions. 

If you have a large number of feature flags, this library may do a lot of work parsing the data and the values. In the future, we may consider only updating if the spreadsheet has changed (via the Google Drive API). I am curious what the level at which this becomes a problem is.

The library internally uses the murmurhash3 algorithm. This is fairly arbitrary but I can't imagine a great argument _against_ it.

We do not support non-string variant values. I can see why it would be reasonable to do so (eg supporting integers), but I think it's a bit of a slippery slope, I have seen some truly horrific abuse of lists, maps, etc in this context. I also don't want to deal with converting types etc, but you can of course do the casting yourself.