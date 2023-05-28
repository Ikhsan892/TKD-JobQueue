package functions

import (
	"assessment/configs"
	"assessment/pkg/adapter/io"
	"assessment/pkg/adapter/postgresql"
	"assessment/pkg/dto"
	"assessment/pkg/entity"
	"assessment/pkg/utils"
	"errors"
	"fmt"
	"github.com/adjust/rmq/v5"
	"strconv"
)

type Volume struct {
	WorkerIndex   int
	Config        *configs.Config
	VolumeRepo    postgresql.IVolumeRepo
	ProjectRepo   postgresql.IProjectRepo
	StructureRepo postgresql.ICompanyStructureRepo
	LogProcess    utils.IJobQueueLog
	processName   string
	errQueue      error
	delivery      rmq.Delivery
}

func NewVolume(worker *Volume) *Volume {
	worker.processName = "Volume Attachment"
	return worker
}

func (v *Volume) error(err error) {
	v.LogProcess.UpdateJobQueueLog(utils.LogInfo{
		ProcessStatus: utils.FAILED,
		ProcessResult: err.Error(),
	})
	utils.Error(v.processName, err)

	errReject := v.delivery.Reject()
	if errReject != nil {
		utils.Error(v.processName, "Error rejecting queue", errReject)
	}

	v.errQueue = err

	defer utils.Info(v.processName, "Job Aborted")
}

func (v *Volume) Consume(delivery rmq.Delivery) {
	payload, errParsePayload := utils.ParsePayload[dto.DownloadAttachment](v.processName, delivery.Payload())
	if errParsePayload != nil {
		v.error(errParsePayload)
	}

	v.delivery = delivery
	utils.Info(v.processName, fmt.Sprintf("Executing Job %s on Worker %d...", "Volume Attachment", v.WorkerIndex))

	v.LogProcess.CreateJobQueueLog(utils.LogInfo{
		ProcessName:    v.processName,
		ProcessPayload: delivery.Payload(),
		ProcessStatus:  utils.PROCESSING,
		ProcessType:    utils.FILE,
		ProcessResult:  "",
		IssuedBy:       payload.UserId,
	})

	var (
		attachmentPath string
		filePath       string
		fileName       string
		volume         entity.GetVolume
		project        entity.GetProject
		structure      entity.CompanyStructure
		err            error
	)

	if payload.VolumeId != nil {
		volume, err = v.VolumeRepo.GetById(*payload.VolumeId)
		if err != nil {
			v.error(err)
		}

		attachments := v.VolumeRepo.GetAttachmentsByVolumeId(*payload.VolumeId)
		if len(attachments) < 1 {
			v.error(errors.New("Attachment is empty"))
		} else {
			attachmentPath = attachments[0].Path
		}

		filePath, fileName = utils.FormatFilePathFormat(v.Config.Report.OutputLocation, "("+strconv.Itoa(int(volume.Id))+")"+volume.StructureId, "zip")
	} else {
		project, err = v.ProjectRepo.GetById(payload.ProjectId)
		if err != nil {
			v.error(err)
		}

		if payload.StructureId == "all" {
			attachmentPath = "/assets/volumes/" + project.Name + "/"
		} else {
			structure, err = v.StructureRepo.GetById(payload.StructureId)
			if err != nil {
				v.error(err)
			}

			attachmentPath = "/assets/volumes/" + project.Name + "/" + structure.Name + "/"
		}

		filePath, fileName = utils.FormatFilePathFormat(v.Config.Report.OutputLocation, project.Name, "zip")
	}

	utils.Debug(v.processName, "Zipping on "+utils.GetTargetZipLocation()+attachmentPath)
	utils.Debug(v.processName, "Zipped result "+filePath)

	zipper := io.NewZip(utils.GetTargetZipLocation()+attachmentPath, filePath)

	if err = zipper.Create(); err != nil {
		v.error(err)
	}

	if v.errQueue == nil {
		err = delivery.Ack()
		if err != nil {
			v.error(err)
		}

		v.LogProcess.UpdateJobQueueLog(utils.LogInfo{
			ProcessName:   fileName,
			ProcessStatus: utils.SUCCESS,
			ProcessResult: filePath,
		})
		utils.Info(v.processName, "Job successfully executed on Worker", v.WorkerIndex)
	}

}
