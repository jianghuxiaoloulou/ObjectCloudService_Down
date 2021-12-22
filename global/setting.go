package global

import (
	"WowjoyProject/ObjectCloudService_Down/pkg/logger"
	"WowjoyProject/ObjectCloudService_Down/pkg/setting"
)

var (
	ServerSetting   *setting.ServerSettingS
	GeneralSetting  *setting.GeneralSettingS
	DatabaseSetting *setting.DatabaseSettingS
	ObjectSetting   *setting.ObjectSettingS
	Logger          *logger.Logger
)
