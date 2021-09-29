import pytest
import redis

from restapi_client import RESTAPI_client

@pytest.fixture()
def setup_restapi_client():
    db = redis.StrictRedis('localhost', 6379, 0)
    db.flushdb()
    cache = redis.StrictRedis('localhost', 6379, 7)
    cache.flushdb()
    configdb = redis.StrictRedis('localhost', 6379, 4)
    configdb.flushdb()
    restapi_client = RESTAPI_client(db)
    restapi_client.post_config_restart_in_mem_db()

    # Sanity check
    keys = db.keys()
    assert keys == []

    keys = cache.keys()
    assert keys == []

    keys = configdb.keys()
    assert keys == []

    yield db, cache, configdb, restapi_client
    
