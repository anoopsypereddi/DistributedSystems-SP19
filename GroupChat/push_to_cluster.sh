for dest in $(<servers.txt); do
  scp mp1.go "$dest:/home/anoops2"
done
