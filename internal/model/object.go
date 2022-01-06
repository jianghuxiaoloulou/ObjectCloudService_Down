package model

import (
	"WowjoyProject/ObjectCloudService_Down/global"
	"WowjoyProject/ObjectCloudService_Down/pkg/general"
	"time"
)

// 通过请求号获取数据Request
func GetRequestData(key string, reqType global.RequestType, actionType global.ActionMode) {
	sql := ""
	switch reqType {
	case global.AccessNumber:
		sql = `select ins.instance_key,ins.file_name,im.img_file_name,fr.dcm_file_name_remote,fr.img_file_name_remote,sl.ip,sl.s_virtual_dir,fr.dcm_file_exist,fr.img_file_exist
		from instance ins
		left join image im on ins.instance_key = im.instance_key
		left join file_remote fr on ins.instance_key = fr.instance_key
		left join study_location sl on sl.n_station_code = ins.location_code
		left join study s on s.study_key = ins.study_key
		where s.accession_number = ?;`
	case global.UidEnc:
		sql = `select ins.instance_key,ins.file_name,im.img_file_name,fr.dcm_file_name_remote,fr.img_file_name_remote,sl.ip,sl.s_virtual_dir,fr.dcm_file_exist,fr.img_file_exist
		from instance ins
		left join image im on ins.instance_key = im.instance_key
		left join file_remote fr on ins.instance_key = fr.instance_key
		left join study_location sl on sl.n_station_code = ins.location_code
		left join study s on s.study_key = ins.study_key
		where s.uid_enc = ?;`
	}
	if sql == "" {
		return
	}
	rows, err := global.DBEngine.Query(sql, key)
	if err != nil {
		global.Logger.Fatal(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		key := KeyData{}
		_ = rows.Scan(&key.instance_key, &key.dcmfile, &key.imgfile, &key.dcmremotekey, &key.jpgremotekey, &key.ip, &key.virpath, &key.dcmstatus, &key.jpgstatus)
		if key.imgfile.String != "" && key.jpgstatus.Int16 == int16(global.FileNotExist) {
			file_path := general.GetFilePath(key.imgfile.String, key.ip.String, key.virpath.String)
			data := global.ObjectData{
				InstanceKey: key.instance_key.Int64,
				FileKey:     key.jpgremotekey.String,
				FilePath:    file_path,
				ActionType:  actionType,
				FileType:    global.JPG,
				Count:       1,
			}
			global.ObjectDataChan <- data
		}
		if key.dcmfile.String != "" && key.dcmstatus.Int16 == int16(global.FileNotExist) {
			file_path := general.GetFilePath(key.dcmfile.String, key.ip.String, key.virpath.String)
			data := global.ObjectData{
				InstanceKey: key.instance_key.Int64,
				FileKey:     key.dcmremotekey.String,
				FilePath:    file_path,
				ActionType:  actionType,
				FileType:    global.DCM,
				Count:       1,
			}
			global.ObjectDataChan <- data
		}
	}
}

// 上传数据后更新数据库
func UpdateUplaod(key int64, filetype global.FileModel, status bool) {
	// 获取更新时时间
	local, _ := time.LoadLocation("Local")
	timeFormat := "2006-01-02 15:04:05"
	curtime := time.Now().In(local).Format(timeFormat)
	switch global.ObjectSetting.OBJECT_Store_Type {
	case global.PublicCloud:
		switch filetype {
		case global.DCM:
			if status {
				global.Logger.Info("***公有云DCM数据上传成功，更新状态***")
				sql := `update file_remote fr set fr.dcm_file_exist_obs_cloud = 1,fr.dcm_location_code_obs_cloud = ?,fr.dcm_update_time_obs_cloud = ? where fr.instance_key = ?;`
				global.DBEngine.Exec(sql, global.ObjectSetting.OBJECT_Upload_Success_Code, curtime, key)
			} else {
				global.Logger.Info("***公有云DCM数据上传失败，更新状态***")
				sql := `update file_remote fr set fr.dcm_file_exist_obs_cloud = 2 where fr.instance_key = ?;`
				global.DBEngine.Exec(sql, key)
			}
		case global.JPG:
			if status {
				global.Logger.Info("***公有云JPG数据上传成功，更新状态***")
				sql := `update file_remote fr set fr.img_file_exist_obs_cloud = 1,fr.img_update_time_obs_cloud = ? where fr.instance_key = ?;`
				global.DBEngine.Exec(sql, curtime, key)
			} else {
				global.Logger.Info("***公有云JPG数据上传失败，更新状态***")
				sql := `update file_remote fr set fr.img_file_exist_obs_cloud = 2 where fr.instance_key = ?;`
				global.DBEngine.Exec(sql, key)
			}
		}
	case global.PrivateCloud:
		switch filetype {
		case global.DCM:
			if status {
				global.Logger.Info("***私有云DCM数据上传成功，更新状态***")
				sql := `update file_remote fr set fr.dcm_file_exist_obs_local = 1,fr.dcm_location_code_obs_local = ?,fr.dcm_update_time_obs_local = ? where fr.instance_key = ?;`
				global.DBEngine.Exec(sql, global.ObjectSetting.OBJECT_Upload_Success_Code, curtime, key)

			} else {
				global.Logger.Info("***私有云DCM数据上传失败，更新状态***")
				sql := `update file_remote fr set fr.dcm_file_exist_obs_local = 2 where fr.instance_key = ?;`
				global.DBEngine.Exec(sql, key)
			}
		case global.JPG:
			if status {
				global.Logger.Info("***私有云JPG数据上传成功，更新状态***")
				sql := `update file_remote fr set fr.img_file_exist_obs_local = 1,fr.img_update_time_obs_local = ? where fr.instance_key = ?;`
				global.DBEngine.Exec(sql, curtime, key)
			} else {
				global.Logger.Info("***私有云JPG数据上传失败，更新状态***")
				sql := `update file_remote fr set fr.img_file_exist_obs_local = 2 where fr.instance_key = ?;`
				global.DBEngine.Exec(sql, key)
			}
		}
	}
}

// 数据下载更新数据库
func UpdateDown(key int64, filetype global.FileModel, status bool) {
	// 获取更新时时间
	local, _ := time.LoadLocation("Local")
	timeFormat := "2006-01-02 15:04:05"
	curtime := time.Now().In(local).Format(timeFormat)
	switch filetype {
	case global.DCM:
		if status {
			global.Logger.Info("***DCM数据下载成功，更新状态***")
			sql := `update file_remote fr set fr.dcm_file_exist = 1,fr.dcm_update_time_retrieve = ? where fr.instance_key = ?;`
			global.DBEngine.Exec(sql, curtime, key)
		} else {
			global.Logger.Info("***DCM数据下载失败，更新状态***")
			sql := `update file_remote fr set fr.dcm_file_exist = 2 where fr.instance_key = ?;`
			global.DBEngine.Exec(sql, key)
		}
	case global.JPG:
		if status {
			global.Logger.Info("***JPG数据下载成功，更新状态***")
			sql := `update file_remote fr set fr.img_file_exist = 1, where fr.instance_key = ?;`
			global.DBEngine.Exec(sql, curtime, key)
		} else {
			global.Logger.Info("***JPG数据下载失败，更新状态***")
			sql := `update file_remote fr set fr.img_file_exist = 2 where fr.instance_key = ?;`
			global.DBEngine.Exec(sql, key)
		}
	}
}
