import time

from openfeature import api
from openfeature.evaluation_context import EvaluationContext

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from gofeatureflag_python_provider.provider_status import ProviderStatus


def test_ccc():
    try:
        goff_provider = GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="http://localhost:1031", data_flush_interval=100
            )
        )

        api.set_provider(goff_provider)
        client = api.get_client(name="test-client")

        start = time.time()
        while goff_provider.get_status() != ProviderStatus.READY:
            time.sleep(0.1)
            if time.time() - start > 5:
                break

        ctx = EvaluationContext(
            targeting_key="d45e303a-38c2-11ed-a261-0242ac120002",
            attributes={
                "email": "john.doe@gofeatureflag.org",
                "firstname": "john",
                "lastname": "doe",
                "anonymous": False,
                "professional": True,
                "rate": 3.14,
                "age": 30,
                "company_info": {"name": "my_company", "size": 120},
                "labels": ["pro", "beta"],
            },
        )

        t = client.get_boolean_details(
            flag_key="bool_targeting_match",
            default_value=False,
            evaluation_context=ctx,
        )
        print("\n" + t.reason)
        t = client.get_boolean_details(
            flag_key="bool_targeting_match",
            default_value=False,
            evaluation_context=ctx,
        )
        print(t.reason)

        time.sleep(10)

        t = client.get_boolean_details(
            flag_key="bool_targeting_match",
            default_value=False,
            evaluation_context=ctx,
        )
        print(t.reason)

        api.shutdown()
    except Exception as e:
        print(e)
        assert False
