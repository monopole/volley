
count=$1
bottom=1
top=30

if [ "$count" == "" ]
then
  count=1
fi
if [ "$count" -lt "$bottom" ]
then
  count=$bottom
elif [ "$count" -gt "$top" ]
then
  count=$top
fi

echo starting $count games
for item in `seq 1 $count`;
do
  volley >& k${item}.txt &
  sleep 3
done    
