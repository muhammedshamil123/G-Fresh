package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"g-fresh/internal/database"
	"g-fresh/internal/model"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf/v2"
)

func SalesReport(c *gin.Context) {
	download := c.Query("download")
	var input model.PlatformSalesReportInput

	// if err := c.BindJSON(&input); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"errors": err.Error()})
	// 	return
	// }
	input.StartDate = c.Query("start_date")
	input.EndDate = c.Query("end_date")
	input.Limit = c.Query("limit")
	input.PaymentStatus = c.Query("payment_status")
	if input.StartDate == "" && input.EndDate == "" && input.Limit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please provide custom date(start & end date), or specify the limit (day, week, month, year)"})
		return
	}

	if input.Limit != "" {
		limits := []string{"day", "week", "month", "year"}
		found := false
		for _, l := range limits {
			if input.Limit == l {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid limit specified, valid options are: day, week, month, year",
			})
			return
		}
		var startDate, endDate string
		switch input.Limit {
		case "day":
			startDate = time.Now().AddDate(0, 0, -1).Format("02-01-2006")
			endDate = time.Now().Format("02-01-2006")
		case "week":
			startDate = time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Format("02-01-2006")
			endDate = time.Now().Format("02-01-2006")
		case "month":
			startDate = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location()).Format("02-01-2006")
			endDate = time.Now().Format("02-01-2006")
		case "year":
			startDate = time.Now().AddDate(-1, 0, 0).Format("02-01-2006")
			endDate = time.Now().Format("02-01-2006")
		}
		println(startDate, endDate)
		result, amount, err := totalSales(startDate, endDate, input.PaymentStatus)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error processing orders"})
			return
		}
		if download == "true" {
			pdfBytes, err := GeneratePDFReport(result, amount, startDate, endDate, input.PaymentStatus)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  false,
					"message": "failed to generate PDF",
					"error":   err.Error(),
				})
				return
			}

			c.Writer.Header().Set("Content-type", "application/pdf")
			c.Writer.Header().Set("Content-Disposition", "inline; filename=salesreport.pdf")
			c.Writer.Write(pdfBytes)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "successfully created sales report",
			"result":  result,
			"amount":  amount,
		})
		return
	}
	result, amount, err := totalSales(input.StartDate, input.EndDate, input.PaymentStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error processing orders",
			"e":     err,
		})
		return
	}

	if download == "true" {
		pdfBytes, err := GeneratePDFReport(result, amount, input.StartDate, input.EndDate, input.PaymentStatus)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "failed to generate PDF",
				"error":   err.Error(),
			})
			return
		}

		c.Writer.Header().Set("Content-type", "application/pdf")
		c.Writer.Header().Set("Content-Disposition", "inline; filename=salesreport.pdf")
		c.Writer.Write(pdfBytes)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully created sales report",
		"result":  result,
		"amount":  amount,
	})
}
func totalSales(start, end, PaymentStatus string) (model.OrderCount, model.AmountInformation, error) {
	var orders []model.Order
	parsedStart, err := time.Parse("02-01-2006", start)
	if err != nil {
		fmt.Println("error parsing Start time: ", err)
		return model.OrderCount{}, model.AmountInformation{}, fmt.Errorf("error parsing Start time: %v", err)
	}
	parsedEnd, err := time.Parse("02-01-2006", end)
	if err != nil {
		fmt.Println("error parsing End time: ", err)
		return model.OrderCount{}, model.AmountInformation{}, fmt.Errorf("error parsing End time: %v", err)
	}
	fstart := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(), 0, 0, 0, 0, time.UTC)
	fend := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(), 23, 59, 59, 999999999, time.UTC)
	startTime := fstart.Format("2006-01-02T15:04:05Z")
	endDate := fend.Format("2006-01-02T15:04:05Z")
	if startTime > endDate {
		fmt.Println("error parsing End time: ", err)
		return model.OrderCount{}, model.AmountInformation{}, errors.New("start date must be before end date")
	}
	if PaymentStatus == "" {
		if err := database.DB.Where("ordered_at BETWEEN ? AND ? ", startTime, endDate).Find(&orders).Error; err != nil {
			fmt.Println("error parsing End time: ", err)
			return model.OrderCount{}, model.AmountInformation{}, errors.New("error fetching orders")
		}
	} else {
		if err := database.DB.Where("ordered_at BETWEEN ? AND ? AND payment_status =?", startTime, endDate, PaymentStatus).Find(&orders).Error; err != nil {
			fmt.Println("error parsing End time: ", err)
			return model.OrderCount{}, model.AmountInformation{}, errors.New("error fetching orders")
		}
	}

	var orderStatusCounts = map[string]int64{
		"PLACED":           0,
		"CONFIRMED":        0,
		"SHIPPED":          0,
		"OUT FOR DELIVERY": 0,
		"DELIVERED":        0,
		"CANCELED":         0,
		"RETURN REQUEST":   0,
		"RETURNED":         0,
	}
	var totalCount int64
	var AccountInformation model.AmountInformation

	for _, order := range orders {
		AccountInformation.TotalCouponDeduction += RoundDecimalValue(order.CouponDiscountAmount)
		AccountInformation.TotalProductOfferDeduction += RoundDecimalValue(order.ProductOfferAmount)
		AccountInformation.TotalAmountBeforeDeduction += RoundDecimalValue(order.FinalAmount)
		var totalRefundAmount float64
		id := strconv.Itoa(int(order.OrderID))
		database.DB.Model(&model.Payment{}).Where("payment_status = ? AND order_id=?", "REFUND", id).Select("SUM(amount) as total_refund").Row().Scan(&totalRefundAmount)
		AccountInformation.TotalSalesRevenue += RoundDecimalValue(order.TotalAmount)
		AccountInformation.TotalAmountAfterDeduction += RoundDecimalValue(order.TotalAmount) + RoundDecimalValue(totalRefundAmount)
		AccountInformation.TotalRefundAmount += RoundDecimalValue(totalRefundAmount)

		var orderItems []model.OrderItem
		if err := database.DB.Where("order_id =?", order.OrderID).Find(&orderItems).Error; err != nil {
			return model.OrderCount{}, model.AmountInformation{}, errors.New("error fetching order items")
		}
		for _, val := range orderItems {
			AccountInformation.TotalProductSold += val.Quantity
			if val.OrderStatus == "CANCELED" || val.OrderStatus == "RETURNED" {
				AccountInformation.TotalProductReturned += val.Quantity
			}
		}
		for _, status := range []string{"PLACED", "CONFIRMED", "SHIPPED", "OUT FOR DELIVERY", "DELIVERED", "CANCELED", "RETURN REQUEST", "RETURNED"} {
			var count int64
			if err := database.DB.Model(&model.OrderItem{}).Where("order_id =? AND order_status =?", order.OrderID, status).Count(&count).Error; err != nil {
				return model.OrderCount{}, model.AmountInformation{}, errors.New("failed to query order items")
			}
			orderStatusCounts[status] += count
			totalCount += count
		}

	}
	AccountInformation.AverageOrderValue = AccountInformation.TotalAmountAfterDeduction / float64(len(orders))
	var userCount int64
	database.DB.Model(&model.User{}).Count(&userCount)
	AccountInformation.TotalCustomers = uint(userCount)

	return model.OrderCount{
		TotalOrder:          uint(totalCount),
		TotalPLACED:         uint(orderStatusCounts["PLACED"]),
		TotalCONFIRMED:      uint(orderStatusCounts["CONFIRMED"]),
		TotalSHIPPED:        uint(orderStatusCounts["SHIPPED"]),
		TotalOUTFORDELIVERY: uint(orderStatusCounts["OUT FOR DELIVERY"]),
		TotalDELIVERED:      uint(orderStatusCounts["DELIVERED"]),
		TotalCANCELED:       uint(orderStatusCounts["CANCELED"]),
		TotalRETURNREQUEST:  uint(orderStatusCounts["RETURN REQUEST"]),
		TotalRETURNED:       uint(orderStatusCounts["RETURNED"]),
	}, AccountInformation, nil
}
func RoundDecimalValue(value float64) float64 {
	multiplier := math.Pow(10, 2)
	return math.Round(value*multiplier) / multiplier
}
func GeneratePDFReport(result model.OrderCount, amount model.AmountInformation, start, end, PaymentStatus string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "Tabloid", "")
	// Add a page
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 20)
	// Title

	pdf.SetTextColor(0, 0, 255)
	pdf.Cell(100, 10, "")
	pdf.Cell(40, 10, "Sales Report")
	pdf.Ln(20)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 12)
	if PaymentStatus != "" {
		pdf.Cell(40, 10, fmt.Sprintf("Payment status: %v", PaymentStatus))
		pdf.Ln(20)
	}

	pdf.Cell(50, 10, fmt.Sprintf("Start Date: %v", start))
	pdf.Cell(50, 10, fmt.Sprintf("End Date: %v", end))
	pdf.Ln(20)

	// Table header with borders
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(50, 10, "Total Order", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, "Placed", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, "Confirmed", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, "Shipped", "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 10, "Out For Delivery", "1", 1, "C", false, 0, "") // '1' at the end moves to a new line

	// Table body with borders
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(50, 10, fmt.Sprintf("%v", result.TotalOrder), "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("%v", result.TotalPLACED), "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("%v", result.TotalCONFIRMED), "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, fmt.Sprintf("%v", result.TotalSHIPPED), "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 10, fmt.Sprintf("%v", result.TotalOUTFORDELIVERY), "1", 1, "C", false, 0, "")

	// Next row of the table
	pdf.Ln(20)
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(70, 10, "Delivered", "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 10, "Cancelled", "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 10, "Return Request", "1", 0, "C", false, 0, "")
	pdf.CellFormat(70, 10, "Returned", "1", 1, "C", false, 0, "")

	// Table body for the second row
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(70, 10, fmt.Sprintf("%v", result.TotalDELIVERED), "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 10, fmt.Sprintf("%v", result.TotalCANCELED), "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 10, fmt.Sprintf("%v", result.TotalRETURNREQUEST), "1", 0, "C", false, 0, "")
	pdf.CellFormat(70, 10, fmt.Sprintf("%v", result.TotalRETURNED), "1", 1, "C", false, 0, "")

	pdf.Ln(30)

	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(100, 10, "Average order value", "1", 0, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(70, 10, fmt.Sprintf("%v", amount.AverageOrderValue), "1", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(100, 10, "Total Products Sold", "1", 0, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(70, 10, fmt.Sprintf("%v", amount.TotalProductSold), "1", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(100, 10, "Total Products Returned", "1", 0, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(70, 10, fmt.Sprintf("%v", amount.TotalProductReturned), "1", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(100, 10, "Total number of customers", "1", 0, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(70, 10, fmt.Sprintf("%v", amount.TotalCustomers), "1", 1, "C", false, 0, "")

	pdf.Ln(30)

	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Total Amount Before Deduction:    %.2f", amount.TotalAmountBeforeDeduction))
	pdf.Ln(20)

	pdf.SetTextColor(255, 0, 0)
	pdf.Cell(40, 10, fmt.Sprintf("Total Coupon Discount Amount:     %.2f", amount.TotalCouponDeduction))
	pdf.Ln(20)

	pdf.Cell(40, 10, fmt.Sprintf("Total Product Discount Amount:    %.2f", amount.TotalProductOfferDeduction))
	pdf.Ln(20)
	pdf.SetTextColor(0, 0, 0)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Total Amount After Deduction:     %.2f", amount.TotalAmountAfterDeduction))
	pdf.Ln(20)

	pdf.Cell(40, 10, fmt.Sprintf("Total Refund Amount:     			%.2f", amount.TotalRefundAmount))
	pdf.Ln(20)

	pdf.SetTextColor(0, 255, 0)
	pdf.Cell(40, 10, fmt.Sprintf("Total Sales Revenue:    			%.2f", amount.TotalSalesRevenue))
	pdf.Ln(20)

	pdf.SetFont("Arial", "", 12)

	var pdfBytes bytes.Buffer
	err := pdf.Output(&pdfBytes)
	if err != nil {
		return nil, err
	}

	return pdfBytes.Bytes(), nil
}

//https://blog.devgenius.io/tutorial-creating-an-endpoint-to-download-files-using-golang-and-gin-gonic-27abbcf75940
