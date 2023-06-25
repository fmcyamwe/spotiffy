# SpotIffy

### yes it's iffy :D

### The Why

Ever been listening to a string of songs and wish that you could just have similar songs to whatever you are currently listening to added to your listening queue?
That is the reason one create playlists right? In order to put those "pump you up" or "romantic" songs in the same place for later.
But that can also get stale after a while, requiring updating and adding new songs to your playlist. My ear immediately picks it up when, at the end of one of my playlists, Spotify thrust onto me one of their 'recommendations' song that is so far off, I curse myself for not having enabled the repeat button beforehand.  

### The What

Using what this gentleman [Klangspektrum](https://github.com/schwamic/klangspektrum2) has done as well as Spotify's [API](https://developer.spotify.com/documentation/web-api/) the original idea was to analyse the previous _x_ number of songs listened to and automatically add to the current queue some related songs, keeping closely with the same mood, genre, etc.

### The How

With every song having audio [features](https://developer.spotify.com/documentation/web-api/reference/get-several-audio-features) I hoped to do a simple averaging or some other calculations for a couple of songs as a test. Afterwards, with a calculated range, I would query Spotify for songs suggestions that fall into it.

Well it is apparently not possible to dynamically add songs to an ongoing queue (_booo_)--unless I missed how to do such a thing. It is possible to create a playlist on the fly though and that is the route taken. 

### And now...

Well....let's just say that the small test I did was inconclusive. A simple average of baseline test songs did not return song recommendation that in my opinion were close. Perhaps this was due to me averaging all the features of the songs instead of limiting this to a smaller set of features. I suspect some audio features play a bigger factor for some type/genre of songs versus a different of genre. 
It would be nice if Spotify would provide this extra information...or at least surface how they rate those audio features in the first place (the inconsistency in rating for two seemingly similar songs was very disconcerting)

### What's next?

The beginning of this pet project was coded in Go but I believe I shall try to code the next iteration using Python and take advantage of it's data analysis libraries. 

At first though would be constructing a playlist with songs that are very much alike for one audio feature. Then using that audio feature as the basis, see the correlation with other audio features that would allow me to reliably construct a closely related playlist using a smaller set of audio features when querying Spotify's API.

Well...that will depend on how much I can tolerate the iffyness :D
