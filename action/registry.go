package action

import "github.com/Sn0wo2/CatSync/config"

var HandlerRegistry = make(map[config.Operation]Handler)
