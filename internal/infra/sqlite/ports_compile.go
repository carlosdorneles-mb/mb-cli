package sqlite

import "mb/internal/ports"

var _ ports.PluginCacheStore = (*Store)(nil)
