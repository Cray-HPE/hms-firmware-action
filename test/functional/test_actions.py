#  MIT License
#
#  (C) Copyright [2020-2022] Hewlett Packard Enterprise Development LP
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
import models

def test_actions():
  print("ACTIONS")
  print("CREATE ACTIONS")
  action = models.ActionParameter(models.StateComponentFilter(),
  models.InventoryHardwareFilter(), models.TargetFilter(),
  models.ImageFilter(), models.Command())
  action_ret = pytest.FAS.create_action(action)
  print("Returned ", action_ret.status_code)
  if action_ret.status_code == 202:
      print(action_ret.json())
      action_json = action_ret.json()
      actionid = action_json["actionID"]
      print(actionid)
  else:
      print("Error Creating Action")
      sys.exit(2)

  print("GET ACTIONS")
  action_ret = pytest.FAS.get_actions()
  print(action_ret.json())

  print("GET ACTION")
  action_ret = pytest.FAS.get_action(actionid)
  if action_ret.status_code == 200:
      print(action_ret.json())
  else:
      print("Error Retrieving Action")
      sys.exit(2)

  print("DELETE ACTION")
  action_ret = pytest.FAS.delete_action(actionid)
  if action_ret.status_code == 204:
      print(action_ret.status_code)
  else:
      print(action_ret.status_code)
      print("Error Deleting Action")
      sys.exit(2)

  print("GET ACTIOn")
  action_ret = pytest.FAS.get_action(actionid)
  if action_ret.status_code == 404:
      print(action_ret.status_code)
      print(action_ret.json())
  else:
      print(action_ret.status_code)
      print("Error Retrieving Deleted Action")
      sys.exit(2)

  print("ACTION TESTS SUCCESSFUL")
