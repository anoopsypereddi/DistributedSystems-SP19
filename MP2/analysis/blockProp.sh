for server in $(seq -f "%02g" 1 10); do
  scp $NETID@sp19-cs425-g54-$server.cs.illinois.edu:/home/$NETID/go/src/mp2/logs/*.transactions . &
done
wait

cat *.transactions | grep BLOCK | awk '{print $2 " " $4}' >> blockTimestamp.dat 

rm *.transactions

