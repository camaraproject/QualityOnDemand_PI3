var doc = {
    "asIpv4Addr": "10.10.1.100",  // This is the asId->ipv4addr to be used by QoD Client while accessing QoD Service. The QoD Service will validate the 'asId' with this provisioned DB
    "scsAsId": "spryfoxnetworks", // Corresponding 'scsAsId' associated with the 'asId'
    "qosMap": {					  // QoD Profile (CAMARA) to QoS Reference (NEF) mapping 
        "QOS_E": "qos-66",
        "QOS_S": "qos-77",
        "QOS_M": "qos-88",
        "QOS_L": "qos-99"
    },
}
printjson(doc)
db.camara.qod.provisionedData.session.insertOne(doc)

// Add more entries as appropriate

