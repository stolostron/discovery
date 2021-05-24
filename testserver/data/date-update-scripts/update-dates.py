# Copyright Contributors to the Open Cluster Management project

import datetime
import os
import json

def update_timestamps(_scenario_json, _yesterday_in_iso, _today_in_iso):
    _subscriptions = _scenario_json["items"]
    for _subs in _subscriptions:
       _subs["created_at"] = _yesterday_in_iso
       _subs["updated_at"] = _today_in_iso
    return _scenario_json

def main():
    _today_in_iso = datetime.datetime.today().isoformat()+"Z"
    _yesterday_in_iso = (datetime.datetime.today() - datetime.timedelta(days=1)).isoformat()+"Z"


    _disco_base_dir = os.getcwd()
    _scenarios_base_dir = os.path.join(_disco_base_dir, "testserver", "data", "scenarios")

    for root, dirs, files in os.walk(_scenarios_base_dir):
        for name in files:
            if name.endswith((".json")):
                _full_scenario_json_path = os.path.join(root, name)
                _scenario_file = open (_full_scenario_json_path, "r")
                _scenario_json = update_timestamps(json.loads(_scenario_file.read()), _yesterday_in_iso, _today_in_iso)
                with open(_full_scenario_json_path, "w") as f:
                    f.write(json.dumps(_scenario_json, indent=4))

main()