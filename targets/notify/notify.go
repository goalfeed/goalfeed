package notify

import "goalfeed/models"

// BroadcastGame is set by the web API at startup to forward game updates to clients.
var BroadcastGame func(models.Game)

// BroadcastGamesList is set by the web API at startup to forward full game lists to clients.
var BroadcastGamesList func()
