import datetime
import sys

if len(sys.argv) <= 1:
    print("No time provided");
    sys.exit(1)

date_time_str = sys.argv[1]
date_time_obj = datetime.datetime.fromisoformat(date_time_str)

print('Date:', date_time_obj.date())
print('Time:', date_time_obj.time())
