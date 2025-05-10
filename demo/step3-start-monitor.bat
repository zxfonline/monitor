echo "go install github.com/divan/expvarmon@latest"
go install github.com/divan/expvarmon@latest

expvarmon -ports="http://localhost:1234/debug/pprof/vars" -vars="mem:memstats.Alloc,Goroutines,tracetotal.Actor1.Count,duration:tracetotal.Actor1.Mean,duration:tracetotal.Actor1.StdDev,duration:tracetotal.Actor1.Median,mem:memstats.Sys,mem:memstats.HeapAlloc,mem:memstats.HeapInuse,memstats.EnableGC,memstats.NumGC,duration:memstats.PauseNs,duration:Uptime,chanstats.chanMainProc.Len,chanstats.chanMainProc.Rate" -i=1s

pause
