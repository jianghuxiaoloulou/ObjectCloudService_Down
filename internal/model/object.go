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
		sql = `select ins.instance_key,ins.file_name,im.img_file_name,sl.ip,sl.s_virtual_dir
		from instance ins
		join image im on ins.instance_key = im.instance_key
		join study_location sl on sl.n_station_code = ins.location_code
		join study s on s.study_key = ins.study_key
		where s.accession_number = ?;`
	case global.UidEnc:
		sql = `select ins.instance_key,ins.file_name,im.img_file_name,sl.ip,sl.s_virtual_dir
		from instance ins
		join image im on ins.instance_key = im.instance_key
		join study_location sl on sl.n_station_code = ins.location_code
		join study s on s.study_key = ins.study_key
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
		_ = rows.Scan(&key.instance_key, &key.dcmfile, &key.imgfile, &key.ip, &key.virpath)
		if key.imgfile.String != "" {
			file_key, file_path := general.GetFilePath(key.imgfile.String, key.ip.String, key.virpath.String)
			data := global.ObjectData{
				InstanceKey: key.instance_key.Int64,
				FileKey:     file_key,
				FilePath:    file_path,
				ActionType:  actionType,
				FileType:    global.JPG,
				Count:       1,
			}
			global.ObjectDataChan <- data
		}
		if key.dcmfile.String != "" {
			file_key, file_path := general.GetFilePath(key.dcmfile.String, key.ip.String, key.virpath.String)
			data := global.ObjectData{
				InstanceKey: key.instance_key.Int64,
				FileKey:     file_key,
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
	local, _ := time.LoadLocation("")
	timeFormat := "2006-01-02 15:04:05"
	curtime := time.Now().In(local).Format(timeFormat)
	switch global.ObjectSetting.OBJECT_Store_Type {
	case global.PublicCloud:
		switch filetype {
		case global.DCM:
			if status {
				global.Logger.Info("***公有云DCM数据上传成功，更新状态***")
				sql := `update instance ins set ins.file_exist_obs_cloud = 1,ins.location_code_obs_cloud = ?,ins.update_time_obs_cloud = ? where ins.instance_key = ?;`
				global.DBEngine.Exec(sql, global.ObjectSetting.OBJECT_Upload_Success_Code, curtime, key)
			} else {
				global.Logger.Info("***公有云DCM数据上传失败，更新状态***")
				sql := `update instance ins set ins.file_exist_obs_cloud = 2 where ins.instance_key = ?;`
				global.DBEngine.Exec(sql, key)
			}
		case global.JPG:
			if status {
				global.Logger.Info("***公有云JPG数据上传成功，更新状态***")
				sql := `update image im set im.file_exist_obs_cloud = 1,im.update_time_obs_cloud = ? where im.instance_key = ?;`
				global.DBEngine.Exec(sql, curtime, key)
			} else {
				global.Logger.Info("***公有云JPG数据上传失败，更新状态***")
				sql := `update image im set im.file_exist_obs_cloud = 2 where im.instance_key = ?;`
				global.DBEngine.Exec(sql, key)
			}
		}
	case global.PrivateCloud:
		switch filetype {
		case global.DCM:
			if status {
				global.Logger.Info("***私有云DCM数据上传成功，更新状态***")
				sql := `update instance ins set ins.file_exist_obs_local = 1,ins.location_code_obs_local = ?,ins.update_time_obs_local = ? where ins.instance_key = ?;`
				global.DBEngine.Exec(sql, global.ObjectSetting.OBJECT_Upload_Success_Code, curtime, key)

			} else {
				global.Logger.Info("***私有云DCM数据上传失败，更新状态***")
				sql := `update instance ins set ins.file_exist_obs_local = 2 where ins.instance_key = ?;`
				global.DBEngine.Exec(sql, key)
			}
		case global.JPG:
			if status {
				global.Logger.Info("***私有云JPG数据上传成功，更新状态***")
				sql := `update image im set im.file_exist_obs_local = 1,im.update_time_obs_local = ? where im.instance_key = ?;`
				global.DBEngine.Exec(sql, curtime, key)
			} else {
				global.Logger.Info("***私有云JPG数据上传失败，更新状态***")
				sql := `update image im set im.file_exist_obs_local = 2 where im.instance_key = ?;`
				global.DBEngine.Exec(sql, key)
			}
		}
	}
}

// 数据下载更新数据库
func UpdateDown(key int64, filetype global.FileModel, status bool) {
	// 获取更新时时间
	local, _ := time.LoadLocation("")
	timeFormat := "2006-01-02 15:04:05"
	curtime := time.Now().In(local).Format(timeFormat)
	switch filetype {
	case global.DCM:
		if status {
			global.Logger.Info("***DCM数据下载成功，更新状态***")
			sql := `update instance ins set ins.FileExist = 1,ins.update_time_retrieve = ? where ins.instance_key = ?;`
			global.DBEngine.Exec(sql, curtime, key)
		} else {
			global.Logger.Info("***DCM数据下载失败，更新状态***")
			sql := `update instance ins set ins.FileExist = 2 where ins.instance_key = ?;`
			global.DBEngine.Exec(sql, key)
		}
	case global.JPG:
		if status {
			global.Logger.Info("***JPG数据下载成功，更新状态***")
			sql := `update image im set im.file_exist = 1,im.update_time_retrieve = ? where im.instance_key = ?;`
			global.DBEngine.Exec(sql, curtime, key)
		} else {
			global.Logger.Info("***JPG数据下载失败，更新状态***")
			sql := `update image im set im.file_exist = 2 where im.instance_key = ?;`
			global.DBEngine.Exec(sql, key)
		}
	}
}
