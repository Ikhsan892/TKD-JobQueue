package functions

import (
	"assessment/configs"
	"assessment/pkg/dto"
	"assessment/pkg/entity"
	"assessment/pkg/utils"
	"errors"
	"fmt"
	"github.com/adjust/rmq/v5"
	"github.com/xuri/excelize/v2"
	"strconv"
	"sync"
)

type IQuestionerRepo interface {
	GetAllQuestion(projectId, templateId uint) []entity.GetQuestion
	GetTemplateQuestionOnProject(projectId uint) []entity.GetTemplate
	GetAllUnitKerja(projectId uint) []entity.GetUnitKerja
	GetProjectDetail(projectId uint) (entity.GetProjectDetail, error)
	GetAnswers(projectId, templateId uint, unitId string) []string
}

type RowValue struct {
	Index        int
	Row          int
	TemplateName string
	Value        string
	Next         func(*excelize.File, string)
}

type Questioner struct {
	WorkerIndex    int
	Config         *configs.Config
	QuestionerRepo IQuestionerRepo
	LogProcess     utils.IJobQueueLog
	processName    string
	errQueue       error
	excel          *excelize.File
	delivery       rmq.Delivery
	logPayload     *utils.LogInfo
	mut            sync.Mutex
}

func NewQuestioner(report *Questioner) *Questioner {
	report.processName = "Report Questioner"
	report.excel = excelize.NewFile()
	return report
}

var (
	additionalHeaders = []entity.GetQuestion{
		{
			Id:        1,
			ProjectId: 0,
			Name:      "No",
		},
		{
			Id:        2,
			ProjectId: 0,
			Name:      "Unit Kerja",
		},
	}
	headerStartRow = 2
	bodyStartRow   = 3
)

func (q *Questioner) terminate(delivery rmq.Delivery) {
	if errExcel := q.excel.Close(); errExcel != nil {
		utils.Error(q.processName, errExcel)
		errReject := delivery.Reject()
		if errReject != nil {
			utils.Error(q.processName, errReject)
		}
	}
}

func (q *Questioner) setValue(data RowValue) error {
	var (
		delivery = q.delivery
		excel    = q.excel
	)

	columnName, errConvNumToName := excelize.ColumnNumberToName(data.Index + 1)
	if errConvNumToName != nil {
		utils.Error(q.processName, errConvNumToName)
		delivery.Reject()
		return errConvNumToName
	}

	var (
		column   = columnName + strconv.Itoa(data.Row)
		colWidth = float64(utils.GetAutoWidth(data.Value))
	)

	excel.SetColWidth(data.TemplateName, columnName, columnName, colWidth)
	excel.SetCellValue(data.TemplateName, column, data.Value)

	data.Next(excel, column)

	return nil
}

func (q *Questioner) baseStyle(templateName, column, color string, bold bool) {
	var (
		excel = q.excel
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
	q.excel.SetCellStyle(templateName, column, column, colHeaderStyle)
}

func (q *Questioner) mapTemplate(projectId uint, next func(entity.GetTemplate, string)) {
	var (
		questionerRepo    = q.QuestionerRepo
		excel             = q.excel
		templateQuestions = questionerRepo.GetTemplateQuestionOnProject(projectId)
	)

	for index, template := range templateQuestions {
		templateName := fmt.Sprintf("%s - %s", "Template", strconv.Itoa(index+1))
		sheetIndex, errCreateNewSheet := excel.NewSheet(templateName)
		if errCreateNewSheet != nil {
			q.error(errCreateNewSheet)
			break
		}

		excel.SetActiveSheet(sheetIndex)

		q.mut.Lock()
		excel.SetCellValue(templateName, "A1", template.Name)
		next(template, templateName)
		q.mut.Unlock()
	}
}

func (q *Questioner) setHeader(payload dto.PayloadQuestioner, wg *sync.WaitGroup) {
	var (
		questionerRepo    = q.QuestionerRepo
		templateQuestions = questionerRepo.GetTemplateQuestionOnProject(payload.ProjectId)
	)

	q.mapTemplate(payload.ProjectId, func(template entity.GetTemplate, templateName string) {
		questions := questionerRepo.GetAllQuestion(payload.ProjectId, template.Id)
		questions = append(additionalHeaders, questions...)

		utils.Debug(q.processName, "template : ", template.Name, " , total_questions : ", len(questions))

		if q.isQuestionAndTemplateEmpty(questions, templateQuestions) {
			q.error(errors.New("questions Empty && template Empty"))
		}

		for index, question := range questions {
			err := q.setValue(RowValue{
				Index:        index,
				Row:          headerStartRow,
				TemplateName: templateName,
				Value:        question.Name,
				Next: func(excel *excelize.File, currentColumn string) {
					if index < 2 {
						q.baseStyle(templateName, currentColumn, "FFED00", false)
					} else {
						q.baseStyle(templateName, currentColumn, "B4E4FF", true)
					}
				},
			})
			if err != nil {
				break
			}
		}

		questions = []entity.GetQuestion{}
	})

	wg.Done()
}

func (q *Questioner) error(err error) {
	q.LogProcess.UpdateJobQueueLog(utils.LogInfo{
		ProcessStatus: utils.FAILED,
		ProcessResult: err.Error(),
	})
	utils.Error(q.processName, err)

	errReject := q.delivery.Reject()
	if errReject != nil {
		utils.Error(q.processName, "Error rejecting queue", errReject)
	}

	q.errQueue = err

	defer utils.Info(q.processName, "Job Aborted")
}

func (q *Questioner) isQuestionAndTemplateEmpty(questions []entity.GetQuestion, templates []entity.GetTemplate) bool {
	return len(questions) < 1 && len(templates) < 1
}

func (q *Questioner) setFilePath(outputLoc, fileName string) (string, string) {
	filePath, fileName := utils.FormatFilePath(outputLoc, fileName)

	utils.Debug(q.processName, filePath)

	q.excel.Path = filePath
	q.excel.DeleteSheet("Sheet1")

	return filePath, fileName
}

func (q *Questioner) setBody(payload dto.PayloadQuestioner, wg *sync.WaitGroup) {
	var (
		questionerRepo = q.QuestionerRepo
		values         [][]string
		rows           []string
	)

	q.mapTemplate(payload.ProjectId, func(template entity.GetTemplate, templateName string) {
		businessUnits := questionerRepo.GetAllUnitKerja(payload.ProjectId)

		for index, unit := range businessUnits {
			answers := questionerRepo.GetAnswers(payload.ProjectId, template.Id, unit.Id)
			rows = append(rows, strconv.Itoa(index+1))
			rows = append(rows, unit.Name)
			rows = append(rows, answers...)

			values = append(values, rows)

			rows = []string{}
		}

		for row, value := range values {
			for i, s := range value {
				err := q.setValue(RowValue{
					Index:        i,
					Row:          row + bodyStartRow,
					TemplateName: templateName,
					Value:        s,
					Next: func(excel *excelize.File, currentColumn string) {
						q.baseStyle(templateName, currentColumn, "ffffff", false)
					},
				})
				if err != nil {
					break
				}
			}
		}

		values = [][]string{}

	})

	wg.Done()
}

func (q *Questioner) Consume(delivery rmq.Delivery) {
	defer q.terminate(delivery)

	q.delivery = delivery
	utils.Info(q.processName, fmt.Sprintf("Executing Job %s on Worker %d...", "Report Questioner", q.WorkerIndex))

	payload, errParsePayload := utils.ParsePayload[dto.PayloadQuestioner](q.processName, delivery.Payload())
	if errParsePayload != nil {
		q.error(errParsePayload)
	}

	q.LogProcess.CreateJobQueueLog(utils.LogInfo{
		ProcessName:    q.processName,
		ProcessPayload: delivery.Payload(),
		ProcessStatus:  utils.PROCESSING,
		ProcessType:    utils.REPORT,
		ProcessResult:  "",
		IssuedBy:       payload.UserId,
	})
	utils.Info(q.processName, "Receiving payload : ", delivery.Payload())

	var wg sync.WaitGroup

	wg.Add(2)

	go q.setHeader(payload, &wg)
	go q.setBody(payload, &wg)

	wg.Wait()

	project, errGetDetail := q.QuestionerRepo.GetProjectDetail(payload.ProjectId)
	if errGetDetail != nil {
		q.error(errGetDetail)
	}

	filePath, fileName := q.setFilePath(q.Config.Report.OutputLocation, fmt.Sprintf("%s(%s)", q.processName, project.Code))

	if q.errQueue == nil {
		if errSave := q.excel.Save(); errSave != nil {
			q.error(errSave)
		}
		delivery.Ack()
		q.LogProcess.UpdateJobQueueLog(utils.LogInfo{
			ProcessName:   fileName,
			ProcessStatus: utils.SUCCESS,
			ProcessResult: filePath,
		})
		utils.Info(q.processName, "Job successfully executed on Worker", q.WorkerIndex)
	}
}
