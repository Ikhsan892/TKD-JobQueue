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

type IInventoryRepo interface {
}

type Inventory struct {
	WorkerIndex   int
	Config        *configs.Config
	ProjectRepo   postgresql.IProjectRepo
	StructureRepo postgresql.ICompanyStructureRepo
	InventoryRepo IInventoryRepo
	LogProcess    utils.IJobQueueLog
	processName   string
	errQueue      error
	excel         *excelize.File
	delivery      rmq.Delivery
	logPayload    *utils.LogInfo
	mut           sync.Mutex
}

func NewInventory(report *Inventory) *Inventory {
	report.processName = "Report Inventory"
	report.excel = excelize.NewFile()
	return report
}

func (i *Inventory) error(err error) {
	i.LogProcess.UpdateJobQueueLog(utils.LogInfo{
		ProcessStatus: utils.FAILED,
		ProcessResult: err.Error(),
	})
	utils.Error(i.processName, err)

	errReject := i.delivery.Reject()
	if errReject != nil {
		utils.Error(i.processName, "Error rejecting queue", errReject)
	}

	i.errQueue = err

	defer utils.Info(i.processName, "Job Aborted")
}

func (i *Inventory) mapTemplate(projectId uint, structureId string, next func(entity.CompanyStructure, string)) {
	var (
		structureRepo = i.StructureRepo
		excel         = i.excel
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
			i.error(errCreateNewSheet)
			break
		}

		excel.SetActiveSheet(sheetIndex)

		i.mut.Lock()
		excel.SetCellValue(templateName, "A1", template.Name)
		next(template, templateName)
		i.mut.Unlock()
	}
}

func (i *Inventory) setValue(data RowValue) error {
	var (
		delivery         = i.delivery
		excel            = i.excel
		columnName       string
		errConvNumToName error
	)
	if data.ColName != "" {
		columnName = data.ColName
	} else {
		columnName, errConvNumToName = excelize.ColumnNumberToName(data.Index + 1)
		if errConvNumToName != nil {
			utils.Error(i.processName, errConvNumToName)
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

func (i *Inventory) baseStyle(templateName, column, color string, bold bool) {
	var (
		excel = i.excel
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
	i.excel.SetCellStyle(templateName, column, column, colHeaderStyle)
}

func (i *Inventory) setHeader(payload dto.ReportVolume, wg *sync.WaitGroup) {
	i.mapTemplate(payload.ProjectId, payload.StructureId, func(template entity.CompanyStructure, templateName string) {

		utils.Debug(i.processName, "template : ", template.Name)

		headers := []string{"ID", "Sarana/Media Simpan", "Unit", "Shelf", "Panjang Shelf", "Volume Sarana Simpan (M)", "Estimasi Volume Media Simpan (%)", "Volume Media Simpan (M)", "Diisi Oleh"}

		for index, header := range headers {
			err := i.setValue(RowValue{
				Index:        index,
				Row:          headerStartRow,
				TemplateName: templateName,
				Value:        header,
				Next: func(excel *excelize.File, currentColumn string) {
					if index < 2 {
						i.baseStyle(templateName, currentColumn, "FFED00", false)
					} else {
						i.baseStyle(templateName, currentColumn, "B4E4FF", true)
					}
				},
			})
			if err != nil {
				i.error(err)
				break
			}
		}
	})

	wg.Done()
}

func (i *Inventory) setFilePath(outputLoc, fileName string) (string, string) {
	filePath, fileName := utils.FormatFilePath(outputLoc, fileName)

	utils.Debug(i.processName, filePath)

	i.excel.Path = filePath
	i.excel.DeleteSheet("Sheet1")

	return filePath, fileName
}

func (i *Inventory) setBody(payload dto.ReportVolume, wg *sync.WaitGroup) {
	var (
	//volumeRepo  = i.InventoryRepo
	//values      [][]interface{}
	//rows        []interface{}
	)

	i.mapTemplate(payload.ProjectId, payload.StructureId, func(template entity.CompanyStructure, templateName string) {
		//values = [][]interface{}{}
	})

	wg.Done()
}

func (i *Inventory) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()

	if err := delivery.Ack(); err != nil {
		panic(err)
	}
}
