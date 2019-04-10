for server in $(seq -f "%02g" 1 10); do
  scp $NETID@sp19-cs425-g54-$server.cs.illinois.edu:/home/$NETID/go/src/mp2/logs/*-$1.snapshot . &
done
wait

echo "digraph snapshot {" > graph-$1.dot
cat *.snapshot >> graph-$1.dot
echo "}" >> graph-$1.dot
dot graph-$1.dot -T png -o graph-$1.png
xdg-open graph-$1.png &
rm *.snapshot
