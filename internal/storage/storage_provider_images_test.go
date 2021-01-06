/*
 * MIT License
 *
 * (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package storage

import (
	"github.com/google/uuid"
)

func (suite *Storage_Provider_TS) Test_Storage_Provider_StoreImage_HappyPath() {
	image := HelperGetStockImage()
	err := MS.StoreImage(image)
	suite.True(err == nil)

	returnImage, err := MS.GetImage(image.ImageID)
	suite.True(err == nil)
	suite.True(returnImage.Equals(image))

	err = MS.DeleteImage(image.ImageID)
	suite.True(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_DeleteImage_NotFound() {
	err := MS.DeleteImage(uuid.New())
	suite.False(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_DeleteImage_Happy() {
	image := HelperGetStockImage()
	err := MS.StoreImage(image)
	suite.True(err == nil)

	err = MS.DeleteImage(image.ImageID)
	suite.True(err == nil)

	// Make sure deleted
	_, err = MS.GetImage(image.ImageID)
	suite.False(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetImage_NotFound() {
	_, err := MS.GetImage(uuid.New())
	suite.False(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetImage_Happy() {
	image := HelperGetStockImage()
	err := MS.StoreImage(image)
	suite.True(err == nil)

	returnImage, err := MS.GetImage(image.ImageID)
	suite.True(err == nil)
	suite.True(returnImage.Equals(image))

	err = MS.DeleteImage(image.ImageID)
	suite.True(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetImages() {
	i1 := HelperGetStockImage()
	i2 := HelperGetStockImage()

	err := MS.StoreImage(i1)
	suite.True(err == nil)

	imageArr, err := MS.GetImages()
	suite.True(err == nil)
	count1 := 0
	count2 := 0
	for _, d := range imageArr {
		if d.ImageID == i1.ImageID {
			suite.True(d.Equals(i1))
			count1++
		}
		if d.ImageID == i2.ImageID {
			suite.True(d.Equals(i2))
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 0)

err = MS.StoreImage(i2)
	suite.True(err == nil)

	imageArr, err = MS.GetImages()
	suite.True(err == nil)
	count1 = 0
	count2 = 0
	for _, d := range imageArr {
		if d.ImageID == i1.ImageID {
			suite.True(d.Equals(i1))
			count1++
		}
		if d.ImageID == i2.ImageID {
			suite.True(d.Equals(i2))
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 1)

	err = MS.DeleteImage(i1.ImageID)
	suite.True(err == nil)
	err = MS.DeleteImage(i2.ImageID)
	suite.True(err == nil)
}
