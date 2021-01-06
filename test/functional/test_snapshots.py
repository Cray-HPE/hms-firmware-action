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
import uuid
import models

def test_snapshots():
  print("SNAPSHOTS")

  n = uuid.uuid1()
  ssname = str(n)
  snapshot = models.SnapshotParameter(ssname, models.StateComponentFilter(),
  models.InventoryHardwareFilter(), models.TargetFilter(), None)

  print("CREATE Snapshot")
  ss_ret = pytest.FAS.create_snapshot(snapshot)
  print("Returned ", ss_ret.status_code)
  if ss_ret.status_code == 201:
      print(ss_ret.json())
  else:
      print("Error Creating Snapshot")
      sys.exit(2)

  print("GET SNAPSHOTS")
  ss_ret = pytest.FAS.get_snapshots()
  print(ss_ret.json())

  print("GET SNAPSHOT")
  ss_ret = pytest.FAS.get_snapshot(ssname)
  if ss_ret.status_code == 200:
      print(ss_ret.json())
  else:
      print("Error Retrieving Snapshot")
      sys.exit(2)

  print("DELETE SNAPSHOT")
  ss_ret = pytest.FAS.delete_snapshot(ssname)
  if ss_ret.status_code == 204:
      print(ss_ret.status_code)
  else:
      print(ss_ret.status_code)
      print("Error Deleting Snapshot")
      sys.exit(2)

  print("GET IMAGE")
  ss_ret = pytest.FAS.get_snapshot(ssname)
  if ss_ret.status_code == 404:
      print(ss_ret.status_code)
      print(ss_ret.json())
  else:
      print(ss_ret.status_code)
      print("Error Retrieving Deleted Snapshot")
      sys.exit(2)

  print("IMAGE TESTS SUCCESSFUL")
