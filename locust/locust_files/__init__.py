from locust_files.db_locust.db_collector import LocustCollector

collector = LocustCollector()
collector.register_collector()
collector.start_http_server()
