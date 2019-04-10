for server in $(seq -f "%02g" 1 10); do
  scp $NETID@sp19-cs425-g54-$server.cs.illinois.edu:/home/$NETID/go/src/mp2/logs/*.transactions . &
done
wait

cat *.transactions | sed 's/$/ 1/' | awk 'NF { a[$1 $2] += $3 } END { for (i in a) print i, a[i] }' | sort > transactions.dat

rm *.transactions

