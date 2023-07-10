package functions

import (
	"assessment/configs"
	"assessment/pkg/adapter/postgresql"
	"assessment/pkg/dto"
	"assessment/pkg/entity"
	"assessment/pkg/utils"
	"fmt"
	"github.com/adjust/rmq/v5"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
	"sync"
)

type VolumeReport struct {
	WorkerIndex   int
	Config        *configs.Config
	VolumeRepo    postgresql.IVolumeRepo
	ProjectRepo   postgresql.IProjectRepo
	StructureRepo postgresql.ICompanyStructureRepo
	LogProcess    utils.IJobQueueLog
	processName   string
	errQueue      error
	excel         *excelize.File
	delivery      rmq.Delivery
	mut           sync.Mutex
}

func NewVolumeReport(worker *VolumeReport) *VolumeReport {
	worker.processName = "Report Volume"
	worker.excel = excelize.NewFile()
	return worker
}

func (v *VolumeReport) error(err error) {
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

func (v *VolumeReport) mapTemplate(projectId uint, structureId string, next func(entity.CompanyStructure, string)) {
	var (
		structureRepo = v.StructureRepo
		excel         = v.excel
		structures    []entity.CompanyStructure
	)

	if structureId != "all" {
		detail, _ := structureRepo.GetById(structureId)
		structures = []entity.CompanyStructure{detail}
	} else {
		structures = structureRepo.GetByProject(projectId)
	}

	for _, template := range structures {
		templateName := fmt.Sprintf("%s", template.Name)
		sheetIndex, errCreateNewSheet := excel.NewSheet(templateName)
		if errCreateNewSheet != nil {
			v.error(errCreateNewSheet)
			break
		}

		excel.SetActiveSheet(sheetIndex)

		v.mut.Lock()
		excel.SetCellValue(templateName, "A1", template.Name)
		next(template, templateName)
		v.mut.Unlock()
	}
}

func (v *VolumeReport) setValue(data RowValue) error {
	var (
		delivery         = v.delivery
		excel            = v.excel
		columnName       string
		errConvNumToName error
	)
	if data.ColName != "" {
		columnName = data.ColName
	} else {
		columnName, errConvNumToName = excelize.ColumnNumberToName(data.Index + 1)
		if errConvNumToName != nil {
			utils.Error(v.processName, errConvNumToName)
			delivery.Reject()
			return errConvNumToName
		}
	}

	var (
		column   = columnName + strconv.Itoa(data.Row)
		colWidth = float64(utils.GetAutoWidth(data.Value))
	)

	excel.SetColWidth(data.TemplateName, columnName, columnName, colWidth)
	excel.SetCellValue(data.TemplateName, column, data.Value)

	if data.MergeCell != "" {
		cells := strings.Split(data.MergeCell, ":")
		excel.MergeCell(data.TemplateName, cells[0], cells[1])
	}

	data.Next(excel, column)

	return nil
}

func (v *VolumeReport) baseStyle(templateName, column, color string, bold bool) {
	var (
		excel = v.excel
	)

	colHeaderStyle, _ := excel.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{color},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{
				Type:  "top",
				Color: "000000",
				Style: 1,
			},
			{
				Type:  "bottom",
				Color: "000000",
				Style: 1,
			},
			{
				Type:  "left",
				Color: "000000",
				Style: 1,
			},
			{
				Type:  "right",
				Color: "000000",
				Style: 1,
			},
		},
		Font: &excelize.Font{
			Bold: bold,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
	})
	v.excel.SetCellStyle(templateName, column, column, colHeaderStyle)
}

func (v *VolumeReport) setHeader(payload dto.ReportVolume, wg *sync.WaitGroup) {

	v.mapTemplate(payload.ProjectId, payload.StructureId, func(template entity.CompanyStructure, templateName string) {

		utils.Debug(v.processName, "template : ", template.Name)

		headers := []string{"ID", "Sarana/Media Simpan", "Unit", "Shelf", "Panjang Shelf", "Volume Sarana Simpan (M)", "Estimasi Volume Media Simpan (%)", "Volume Media Simpan (M)", "Diisi Oleh"}

		for index, header := range headers {
			err := v.setValue(RowValue{
				Index:        index,
				Row:          headerStartRow,
				TemplateName: templateName,
				Value:        header,
				Next: func(excel *excelize.File, currentColumn string) {
					if index < 2 {
						v.baseStyle(templateName, currentColumn, "FFED00", false)
					} else {
						v.baseStyle(templateName, currentColumn, "B4E4FF", true)
					}
				},
			})
			if err != nil {
				v.error(err)
				break
			}
		}
	})

	wg.Done()
}

func (v *VolumeReport) setFilePath(outputLoc, fileName string) (string, string) {
	filePath, fileName := utils.FormatFilePath(outputLoc, fileName)

	utils.Debug(v.processName, filePath)

	v.excel.Path = filePath
	v.excel.DeleteSheet("Sheet1")

	return filePath, fileName
}

func (v *VolumeReport) setBody(payload dto.ReportVolume, wg *sync.WaitGroup) {
	var (
		volumeRepo  = v.VolumeRepo
		values      [][]interface{}
		statValues  [][]interface{}
		rows        []interface{}
		statRows    []interface{}
		statsHeader = []string{"Sarana Simpan", "Total Sarana Simpan (M)", "Total Volume Media Simpan (M)", "Total Estimasi Volume Media Simpan (%)"}
	)

	v.mapTemplate(payload.ProjectId, payload.StructureId, func(template entity.CompanyStructure, templateName string) {
		volumes := volumeRepo.GetByStructureId(template.UUID, payload.ProjectId)
		volumeStats := volumeRepo.GetVolumeStats(template.UUID, payload.ProjectId)

		if len(volumes) > 0 {
			for _, volume := range volumes {
				rows = append(rows, strconv.Itoa(int(volume.Id)))
				rows = append(rows, volume.StorageFacilityName)
				rows = append(rows, strconv.Itoa(int(volume.Unit)))
				rows = append(rows, strconv.Itoa(int(volume.Shelf)))
				rows = append(rows, volume.ShelfLong)
				rows = append(rows, volume.VolumeStorageFacility)
				rows = append(rows, fmt.Sprintf("%.2f %%", float64(volume.VolumeStorageMediaPercentage)))
				rows = append(rows, float64(volume.VolumeStorageMedia))
				rows = append(rows, volume.FilledByName)

				values = append(values, rows)

				rows = []interface{}{}
			}
			totalFacility := 0.00
			totalMedia := 0.00
			totalEstimationVolumePercentage := 0.00
			for _, stat := range volumeStats {
				statRows = append(statRows, stat.Type)
				statRows = append(statRows, stat.TotalFacilityVolume)
				statRows = append(statRows, stat.TotalMediaVolume)

				if stat.TotalFacilityVolume > 0 {
					statRows = append(statRows, fmt.Sprintf("%.2f %%", stat.TotalMediaVolume/stat.TotalFacilityVolume*100))
				} else {
					statRows = append(statRows, fmt.Sprintf("%d %%", 0))
				}

				totalFacility += float64(stat.TotalFacilityVolume)
				totalMedia += float64(stat.TotalMediaVolume)
				statValues = append(statValues, statRows)

				statRows = []interface{}{}
			}
			if totalFacility > 0 {
				totalEstimationVolumePercentage = totalMedia / totalFacility * 100
			}

			statRows = append(statRows, "Grand Total")
			statRows = append(statRows, totalFacility)
			statRows = append(statRows, totalMedia)
			statRows = append(statRows, fmt.Sprintf("%.2f %%", totalEstimationVolumePercentage))
			statValues = append(statValues, statRows)

			for row, value := range values {
				for i, s := range value {
					err := v.setValue(RowValue{
						Index:        i,
						Row:          row + bodyStartRow,
						TemplateName: templateName,
						Value:        s,
						Next: func(excel *excelize.File, currentColumn string) {
							v.baseStyle(templateName, currentColumn, "ffffff", false)
						},
					})
					if err != nil {
						break
					}
				}
			}

			for i, s := range statsHeader {
				v.setValue(RowValue{
					Index:        i,
					Row:          len(values) + 2 + bodyStartRow,
					TemplateName: templateName,
					Value:        s,
					Next: func(excel *excelize.File, currentColumn string) {
						if i < 1 {
							v.baseStyle(templateName, currentColumn, "FFED00", false)
						} else {
							v.baseStyle(templateName, currentColumn, "B4E4FF", true)
						}
					},
				})
			}

			for rowStat, statValue := range statValues {
				for i, s := range statValue {
					err := v.setValue(RowValue{
						Index:        i,
						Row:          len(values) + 3 + bodyStartRow + rowStat,
						TemplateName: templateName,
						Value:        s,
						Next: func(excel *excelize.File, currentColumn string) {
							if rowStat == len(statValues)-1 && i == 0 {
								v.baseStyle(templateName, currentColumn, "FFED00", true)
							} else if i < 1 && rowStat < len(statValues)-1 {
								v.baseStyle(templateName, currentColumn, "B4E4FF", false)
							} else {
								v.baseStyle(templateName, currentColumn, "ffffff", true)
							}
						},
					})
					if err != nil {
						break
					}
				}
			}
			statRows = []interface{}{}

		}

		values = [][]interface{}{}
		statValues = [][]interface{}{}

	})

	wg.Done()
}

func (v *VolumeReport) Consume(delivery rmq.Delivery) {
	payload, errParsePayload := utils.ParsePayload[dto.ReportVolume](v.processName, delivery.Payload())
	if errParsePayload != nil {
		v.error(errParsePayload)
	}

	v.delivery = delivery
	utils.Info(v.processName, fmt.Sprintf("Executing Job %s on Worker %d...", "Volume Report", v.WorkerIndex))

	v.LogProcess.CreateJobQueueLog(utils.LogInfo{
		ProcessName:    v.processName,
		ProcessPayload: delivery.Payload(),
		ProcessStatus:  utils.PROCESSING,
		ProcessType:    utils.FILE,
		ProcessResult:  "",
		IssuedBy:       payload.UserId,
	})

	var wg sync.WaitGroup

	wg.Add(2)

	go v.setHeader(payload, &wg)
	go v.setBody(payload, &wg)

	wg.Wait()

	project, errGetDetail := v.ProjectRepo.GetById(payload.ProjectId)
	if errGetDetail != nil {
		v.error(errGetDetail)
	}

	filePath, fileName := v.setFilePath(v.Config.Report.OutputLocation, fmt.Sprintf("%s(%s)", v.processName, project.Code))

	if v.errQueue == nil {
		if errSave := v.excel.Save(); errSave != nil {
			v.error(errSave)
		}
		delivery.Ack()
		v.LogProcess.UpdateJobQueueLog(utils.LogInfo{
			ProcessName:   fileName,
			ProcessStatus: utils.SUCCESS,
			ProcessResult: filePath,
		})
		utils.Info(v.processName, "Job successfully executed on Worker", v.WorkerIndex)
	}
}
