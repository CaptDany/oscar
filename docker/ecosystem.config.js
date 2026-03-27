{
  "apps": [
    {
      "name": "opencrm",
      "script": "./opencrm",
      "instances": "max",
      "exec_mode": "cluster",
      "env": {
        "NODE_ENV": "production",
        "APP_ENV": "production"
      },
      "error_file": "./logs/error.log",
      "out_file": "./logs/out.log",
      "log_date_format": "YYYY-MM-DD HH:mm:ss Z",
      "merge_logs": true,
      "wait_ready": true,
      "listen_timeout": 8000,
      "kill_timeout": 5000,
      "max_memory_restart": "500M"
    }
  ]
}
