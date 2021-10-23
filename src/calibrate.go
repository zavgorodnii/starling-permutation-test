// Copyright D. Krylov (github.com/dmkrylov)
package src

import (
"github.com/tealeg/xlsx"
"github.com/pkg/errors"
"log"
)

func Calibrate(p float64, NCPath string) (float64, error){
  var calibratedP float64

	if p>0.1{
    calibratedP, err:= CountRatio(p, NCPath)
    if err != nil {
      return 0, errors.Wrapf(err, "failed to count ratio %s", NCPath)
    }
		return calibratedP, nil
	} else {
    calibratedP = 2.59*p-5.89*p*p
    log.Printf("\nCalibrated Max P(costs) = 2.59 * %f - 5.89 * %f^2 = %f", p,p,calibratedP)
    return  calibratedP, nil
	}
}

func CountRatio(p float64, NCPath string)(float64, error){
  var ratio float64
  NCFile, err := xlsx.OpenFile(NCPath)
  if err != nil {
		return 0, errors.Wrapf(err, "failed to read %s", NCPath)
	}

  nums:=[]float64{}
  for idx := 0; idx < len(NCFile.Sheets[0].Rows); idx++ {
    num, err := NCFile.Sheets[0].Rows[idx].Cells[1].Float()
    if err != nil {
			return 0, errors.Wrapf(err, "number from row %d is not a float value", idx)
		}
    nums = append(nums, num)
  }
  var countLessThan float64
  countLessThan = 0
  for idx :=1; idx < len(nums); idx++ {
  if nums[idx] <= p {
    countLessThan ++
  }
}
numsLength := float64(len(nums))
ratio = countLessThan/numsLength
log.Printf("\nCalibrated Max P(costs) = %f / %f = %f", countLessThan,numsLength,ratio)
return ratio, nil
}
