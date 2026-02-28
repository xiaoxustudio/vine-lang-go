# 分析单个文件的CPU和内存使用
sh dev.sh --cpuprofile=cpu.prof --memprofile=mem.prof examples/time/index.vine

# 查看CPU分析结果
go tool pprof cpu.prof

# 查看内存分析结果
go tool pprof mem.prof