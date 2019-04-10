for server in $(seq -f "%02g" 1 10); do
  scp $NETID@sp19-cs425-g54-$server.cs.illinois.edu:/home/$NETID/go/src/mp2/logs/*.transactions . &
done
wait

echo "digraph snapshot {" > blockHash.dot
cat *.transactions | grep BLOCK | awk '{print $2 " -> " $3}' >> blockHash.dot
echo "}" >> blockHash.dot
dot blockHash.dot -T png -o blockHash.png
open blockHash.png &

rm *.transactions

