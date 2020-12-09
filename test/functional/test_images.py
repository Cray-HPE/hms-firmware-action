# MIT License
#
# (C) Copyright [2021] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

import pytest
import sys
import os
import models

def test_images():
  print("IMAGE")
  image = models.Image("NodeBMC","cray",["WNC"],"BMC","s3:test","nc-1.3.5","1.3.5")
  print("CREATE IMAGE")
  image_ret = pytest.FAS.create_image(image)
  print("Returned ", image_ret.status_code)
  if image_ret.status_code == 200:
      print(image_ret.json())
      image_json = image_ret.json()
      imageid = image_json["imageID"]
      print(imageid)
  else:
      print("Error Creating Image")
      sys.exit(2)

  print("GET IMAGES")
  image_ret = pytest.FAS.get_images()
  print(image_ret.json())

  print("GET IMAGE")
  image_ret = pytest.FAS.get_image(imageid)
  if image_ret.status_code == 200:
      print(image_ret.json())
  else:
      print("Error Retrieving Image")
      sys.exit(2)

  print("DELETE IMAGE")
  image_ret = pytest.FAS.delete_image(imageid)
  if image_ret.status_code == 204:
      print(image_ret.status_code)
  else:
      print(image_ret.status_code)
      print("Error Deleting Image")
      sys.exit(2)

  print("GET IMAGE")
  image_ret = pytest.FAS.get_image(imageid)
  if image_ret.status_code == 404:
      print(image_ret.status_code)
      print(image_ret.json())
  else:
      print(image_ret.status_code)
      print("Error Retrieving Deleted Image")
      sys.exit(2)

  print("IMAGE TESTS SUCCESSFUL")
