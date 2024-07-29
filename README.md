# kong-plugin-header

For use you can follow this curl

```curl
curl -i -X POST http://localhost:8001/services/{service_name_or_id_service}/plugins \
  --data "name=kong-plugin-header" \
  --data "config.target_field=Data" \
  --data 'config.key_value_pairs[0].key=payor_code' \
  --data 'config.key_value_pairs[0].value_source=header' \
  --data 'config.key_value_pairs[1].key=corp_code' \
  --data 'config.key_value_pairs[1].value_source=query'
```

change the service name or id service name with service usage this plugin

```code
target_field
```
you can change with json field example :
```json
 {
    "test":"123"
 }
```

if the <b>target_field</b> fill with data the result should be 
```json
 {
    "test":"123",
    "Data" : {
        "payor_code" : "XXX",
        "corp_code" : "xxxx"
    }
 }
```

the result depends on key and value from <b> value_source </b>
