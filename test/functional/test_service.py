#  MIT License
#
#  (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
#
#  Permission is hereby granted, free of charge, to any person obtaining a
#  copy of this software and associated documentation files (the "Software"),
#  to deal in the Software without restriction, including without limitation
#  the rights to use, copy, modify, merge, publish, distribute, sublicense,
#  and/or sell copies of the Software, and to permit persons to whom the
#  Software is furnished to do so, subject to the following conditions:
#
#  The above copyright notice and this permission notice shall be included
#  in all copies or substantial portions of the Software.
#
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
#  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
#  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
#  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
#  OTHER DEALINGS IN THE SOFTWARE.

import pytest
import sys
import os


def test_service():
  print("SERVICE")

  print ("STATUS")
  service_ret = pytest.FAS.get_service_status()
  if service_ret.status_code == 200:
      print(service_ret.json())
  else:
      print("Error Service Status")
      sys.exit(2)

  print ("STATUS DETAILS")
  service_ret = pytest.FAS.get_service_status_details()
  if service_ret.status_code == 200:
      print(service_ret.json())
  else:
      print("Error Service Status Details")
      sys.exit(2)

  print ("VERSION")
  service_ret = pytest.FAS.get_service_version()
  if service_ret.status_code == 200:
      print(service_ret.json())
  else:
      print("Error Service Version")
      sys.exit(2)

  print("SERVICE TESTS SUCCESSFUL")
