package sqlite

import "mb/internal/ports"

var _ ports.PluginSyncStore = (*Store)(nil)
