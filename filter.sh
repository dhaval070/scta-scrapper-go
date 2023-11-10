for ff in `ls var/20231108/*.csv`; do
    # echo "a.csv" |sed "s/.csv//"                                                                                                                                                               main [c85fa24] deleted modified 
    n=`basename "$ff"`
    n=`echo "$n" | sed "s/.csv//"`
    go run ./cmd/site-schedule/main.go -site $n -infile var/20231108/$n.csv > var/20231108/with-surface1/$n.csv
done
