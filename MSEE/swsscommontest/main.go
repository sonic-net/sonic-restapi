package main

import "swsscommon"

func main() {
    db := swsscommon.NewDBConnector(0, "localhost", 6379, 0);
    defer db.Delete()
    pt := swsscommon.NewProducerStateTable(db, "QOS_TABLE")
    defer pt.Delete()
    pt.Del("PORT_TABLE:ETHERNET4", "DEL", "")
    pt.Set("SCHEDULER_TABLE:SCAVENGER", map[string]string{
        "algorithm": "DWRR",
        "weight": "35",
    }, "SET", "")
}
