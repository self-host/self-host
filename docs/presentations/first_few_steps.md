% Self-host API: First steps
% Mikael Ganehag Brorsson @ganehag
% 2021-06-24

# Before venturing forth

Please make sure that you've completed all the necessary steps outlined in [Five to fifteen-minute deployment](https://github.com/self-host/self-host/blob/main/docs/test_deployment.md) walkthrough.

---

Throughout this guide, we will use the Swagger UI interface embedded in the Self-host API server.

Go ahead and access it now.

Authenticate with any user you like, but the user requires **full access** to the entire system. For simplicity's sake, use the root user if you are using your local development environment.

*You remember to change the server URL and add your credentials via the authorize button, right!?*

---

![You've encountered a mysterious text.][dangerous-to-go-alone]


# Create your first time-series

Let's start with something simple. Let's create a time series to store a temperature value.

---

1. Scroll down to the `timeseries` section and expand the green <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/AddTimeSeries">POST /v2/timeseries</a> field.
2. Click on the *Try it out* button.

---

3. Start by erasing the line;

    ```json
     "thing_uuid": "e21ae595-15a5-4f11-8992-9d33600cc1ee",
    ```

4. We don't want any tags added to the time series at this point, so also erase the lines;

    ```json
    "tags": [
      "GT31",
      "ODT"
    ],
    ```
    Or make the item an empty list.

    ```json
    "tags": [],
    ```

---

5. You should now have something that looks like this:

    ```json
    {
      "name": "My Time Series",
      "si_unit": "C",
      "lower_bound": -50,
      "upper_bound": 50
    }
    ```

6. Click on the blue `Execute` button.

---

7. Scroll down to the response section. There should be a response looking similar to this, yet with another UUID.

    ```json
    {
      "created_by": "00000000-0000-1000-8000-000000000000",
      "lower_bound": -50,
      "name": "My Time Series",
      "si_unit": "C",
      "tags": [],
      "thing_uuid": null,
      "upper_bound": 50,
      "uuid": "6ee20313-67b7-4203-8c02-e882fe454fc3"
    }
    ```

---

8. Take note of the UUID in the JSON response, as we will need it in the future.

![You received a new UUID!][you-received-an-item]

---

## Adding data to a time series

Now, let's add a few data points to our new time series.

---

1. Scroll down to the <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/AddDataToTimeseries">POST /v2/timeseries/{uuid}/data</a>field and expand it.

2. Click on the *Try it out* button.

---

3. Erase the example UUID in the field and insert our time series UUID.

4. Take note of the request body format. It's a list of objects where `v` stands for value and `ts` stands for the timestamp. The format for the timestamp is [RFC-3339](https://datatracker.ietf.org/doc/html/rfc3339).

    ```json
    [
      {
        "v": 3.14,
        "ts": "2021-06-22T13:20:55.286Z"
      }
    ]
    ```

5. Click on the blue `Execute` button.

---

6. Scroll down to the response section. As you can see, we got nothing more than an HTTP 201 (Created) response from the server. This response means that everything is OK and that the server stored the data in the database.

7. Let's scroll back up and click on the blue `Execute` button again.

---

8. The response we got this time is not 201 any longer, but 400 with the text: "*Request caused an error due to duplicate key violation*".

    This error means that a data point already exists at that particular point in time for this time series.  

    Two data points can occupy the same point in time. If you want to overwrite a data point, you first need to erase it.

---

9. Let's change the timestamp by adding 1 minute to it. In the example above, we change from `20` to `21`. Then press the blue `Execute` button again.
10. You will get an HTTP 201 (Created) response from the server this time.

---

11. You can repeat step 9 and 10 as many times as you like to add more data points, or you can expand the list of data points and submit more than one at a time, play around with it.

---

## Retrieve data from a time series

With some data point in our time series, we can now query data from it.

---

1. Navigate to the blue <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/QueryTimeseriesForData">GET /v2/timeseries/{uuid}/data</a> field, located right above the field from the last section and expand it.

2. Click on the *Try it out* button.

---

3. Erase the example UUID in the field and insert our time series UUID.

4. Change the `start` and `end` fields to the period we want to query. Again, we are keeping in mind that it can not exceed one year or 365.25 days.

5. Leave the remaining input fields on their respective default value.

6. Click on the blue `Execute` button.

---

7. Scroll down to the response section. There should be a similar response, yet with different timestamps, possibly different values and numbers of items.

    ```json
    [
      {
        "ts": "2021-06-22T13:20:55.286Z",
        "v": 3.14
      },
      {
        "ts": "2021-06-22T13:21:55.286Z",
        "v": 3.14
      },
      {
        "ts": "2021-06-22T13:22:55.286Z",
        "v": 3.14
      }
    ]
    ```

---

8. The query interface supports different operations to apply to the data before creating a response. Such operations include;
    - *unit* conversion
    - *greater or equal* and *less or equal* filtering
    - *aggregation* using *precision truncation* of the timestamp and time zone declaration.

---

9. Let's see what happens if we change the requested unit from C(elsius) to F(ahrenheit). Do this by replacing the C in the unit field with an F. Then click on the blue `Execute` button to perform the request once more.

---

10. Scroll down to the response section. There should be a response looking similar to this, yet with (possibly) different elements. Moreso, as you can see, the value is not the same as before; it is converted from Celsius to Fahrenheit.

    ```json
    [
      {
        "ts": "2021-06-22T13:20:55.286Z",
        "v": 37.652
      },
      {
        "ts": "2021-06-22T13:21:55.286Z",
        "v": 37.652
      },
      {
        "ts": "2021-06-22T13:22:55.286Z",
        "v": 37.652
      }
    ]
    ```

---

11. Let's try with something else. Maybe `kWh`. What will happen?

---

12. You will get an HTTP 400 error with the message "*Unable to convert to the requested unit*" because converting from C to kWh is impossible.
    
    They are not compatible.

---

13. Change back to C for the unit and set the `precision` to **hour** and the `aggregate` to `count`.

14. The result you get depends on the number of data points in the giver period and the range of the period.

    We truncate each timestamp on the hour, then perform a `count` operation on the result. As you might have surmised already, we are left with the number of data points for an hour.

---

## Delete data from a time series

For the final step in this segment, let's remove one of our data points.

---

1. Scroll down to the red <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/DeleteDataFromTimeSeries">DELETE /v2/timeseries/{uuid}/data</a> field and expand it.

2. Click on the *Try it out* button.

---

3. Erase the example UUID in the field and insert our time series UUID.

4. Change the `start` and `end` fields to a period so that only a single data point is within it.

    Leave the `ge` and `le` fields empty.

5. Click on the blue `Execute` button.

---

6. Scroll down to the response section. As you can see, we got nothing more than an HTTP 204 (No Content) response from the server. This response means that everything is OK and that the server removed any data in the range from the database.

7. If you want to verify this, you can go back to the previous section and perform a new query over the same period; you shall have one data point less than before.

---

Great work!

![On to the next section!][onwards-and-upwards]


# Creating a Thing

A thing is a collection of time series, representing... anything.

---

1. Scroll up to the green <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/things/AddThing">POST /v2/things</a> field and expand it.

2. Click on the *Try it out* button.

---

3. Remove the "tags" and "type" fields from the example request body.

    Giving you a request that looks like this;

    ```json
    {
      "name": "My Thing"
    }
    ```

4. Click on the blue `Execute` button.

---

5. Scroll down to the response section. There should be a response looking similar to this, yet with another UUID.

    ```json
    {
      "created_by": "00000000-0000-1000-8000-000000000000",
      "name": "My Thing",
      "state": "inactive",
      "tags": [],
      "type": null,
      "uuid": "20b9915d-9208-4a4a-8ab0-6fa45beb8edf"
    }
    ```

    *This UUID is also important, so keep track of it!*

---

#### I see that the state is inactive. What does this mean?

At the moment, nothing. There are four different states;

active, inactive, passive and archived

They are intended as signalling states, used by internal and external services when interacting with the system.


## Assigning a time series to a Thing

We now have one time series, and one thing, let's connect them.

---

1. Scroll down to the orange <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/UpdateTimeseriesByUuid">PUT /v2/timeseries/{uuid}</a> field and expand it.

2. Click on the *Try it out* button.

---

3. Erase the example UUID in the field and insert our time series UUID.

4. Edit the request body so that it looks like this;

    ```json
    {
      "thing_uuid": "20b9915d-9208-4a4a-8ab0-6fa45beb8edf"
    }
    ```

    Then edit the thing_uuid so that it matches the UUID from your Thing.

5. Click on the blue `Execute` button.

---

6. Scroll down to the response section. As you can see, we got nothing more than an HTTP 204 (No Content) response from the server. This response means that everything is OK and that the server has updated the target.


## List time series assigned to a Thing

With our time series assigned to a thing, let's query the thing for a list of time series.

---

1. Scroll up to the blue <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/FindTimeSeriesForThing">GET /v2/things/{uuid}/timeseries</a> field and expand it.

2. Click on the *Try it out* button.

3. Erase the example UUID in the field and insert our thing UUID.

4. Click on the blue `Execute` button.

---

5. Scroll down to the response section. The response you get shall look similar to this;

    ```json
    [
      {
        "created_by": "00000000-0000-1000-8000-000000000000",
        "lower_bound": -50,
        "name": "My Time Series",
        "si_unit": "C",
        "tags": [],
        "thing_uuid": "20b9915d-9208-4a4a-8ab0-6fa45beb8edf",
        "upper_bound": 50,
        "uuid": "6ee20313-67b7-4203-8c02-e882fe454fc3"
      }
    ]
    ```

---

We can query all time series assigned to a specific thing via this interface.

Feel free to peruse the API documentation on your own time and learn about all the possible interactions.

In short, the purpose of a thing is to act as a hub for time series and datasets. We are making it easier to keep track of ... everything.

---

Still with us? Great!

![On to the next section!][onwards-and-upwards]


# Create a dataset

A dataset is a "complex" data structure. Data that doesn't fit into time series.

---

Examples are;

- Computed, filtered, aggregated and mangled data.
    + CSV files.
    + XML files.
    + etc.
- Configuration files:
    + YAML.
    + JSON.
    + etc.

---

Datasets are in themselves not used by the API server. They are used by external services and programs to handle dynamic configuration and store smaller data sets.

---

1. Scroll up to the green <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/datasets/AddDatasets">POST /v2/datasets</a> field and expand it.

2. Click on the *Try it out* button.

---

3. Edit the request body by changing the thing_uuid value to the UUID of our previously created thing, also erase the tags. Leave the rest as is.

    ```{.json data-line-numbers="|2" data-id="code-animation"}
    {
      "content": "aGVsbG8sIHdvcmxkIQ==",
      "format": "ini",
      "name": "foo",
      "thing_uuid": "20b9915d-9208-4a4a-8ab0-6fa45beb8edf"
    }
    ```

4. Click on the blue `Execute` button.

---

5. Scroll down to the response section. The response you get shall look similar to this;

    ```{.json data-line-numbers="|2|3-4|5|7|10-11|12" data-id="code-animation"}
    {
      "checksum": "68e656b251e67e8358bef8483ab0d51c6619f3e7a1a9f0e75838d41ff368f728",
      "created": "2021-06-24T08:07:49.182709+02:00",
      "created_by": "00000000-0000-1000-8000-000000000000",
      "format": "ini",
      "name": "foo",
      "size": 13,
      "tags": [],
      "thing_uuid": "20b9915d-9208-4a4a-8ab0-6fa45beb8edf",
      "updated": "2021-06-24T08:07:49.182709+02:00",
      "updated_by": "00000000-0000-1000-8000-000000000000",
      "uuid": "05b8e64e-e339-4d3b-adbf-9d32b762d3b0"
    }
    ```

*This UUID is also important, so keep track of it!*

---

## Retrieving the content of a dataset

With our data set created. Let's find out what content it is hiding.

---

1. Scroll up to the blue <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/datasets/GetRawDatasetByUuid">GET /v2/datasets/{uuid}/raw</a> field and expand it.

2. Click on the *Try it out* button.

3. Erase the example UUID in the field and insert our dataset UUID. Also, erase the If-None-Match value. We will revisit this.

4. Click on the blue `Execute` button.

---

5. Scroll down to the response section. The response you get shall look like this;

    ```text
    hello, world!
    ```

    The Content-Type is also set to "text/plain; charset=utf-8" in the response header. Different formats result in a response with different Content-Type.

---


## Updating the dataset

Let's change the content and format of our dataset to something else.

---

1. Scroll up to the orange <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/datasets/UpdateDatasetByUuid">PUT /v2/datasets/{uuid}</a> field and expand it.

2. Click on the *Try it out* button.

---

3. Erase the example UUID in the field and insert our dataset UUID.

4. Erase the name and tags field so that you are left with the following;

    ```json
    {
      "content": "LS0tCiBkb2U6ICJhIGRlZXIsIGEgZmVtYWxlIGRlZXIiCiBwaTogMy4xNDE1OQo=",
      "format": "yaml"
    }
    ```
5. Click on the blue `Execute` button.

---

6. Scroll down to the response section. As you can see, we got nothing more than an HTTP 204 (No Content) response from the server. This response means that everything is OK and that the server has updated the target.

---

7. Scroll back to the blue <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/datasets/GetRawDatasetByUuid">GET /v2/datasets/{uuid}/raw</a> field.

8. Execute the same query as before.

---

9. Scroll down to the response section. The response you get shall now look like this;

    ```yaml
    ---
    doe: "a deer, a female deer"
    pi: 3.14159
    ```

    Instead of "hello, world!" we have something else, and the Content-Type has changed to "application/yaml".

---

10. From the response header section, copy the ETag value.

    ```text
    a4d298431f2435991ab4c47c49e14627447b643bb1702b263dad8a49060db384
    ```

11. Input this value into the If-None-Match field.

12. Execute the query once more.

---

13. Instead of the dataset's content, we get a 304 (Not Modified) error. This status tells us that by comparison of checksums that the file's content has not changed.

    This reply is helpful to clients so that they can keep a local cache of the dataset content. By submitting the known checksum via If-None-Match, we can avoid having to return the entire dataset content.

---


This section was a short one.

![On to the next section!][onwards-and-upwards]


# Managing tags

Tags allow us to assign keys to time series, things, datasets and program. This aids with tracking.

---

Tags work in the same way wherever they are supported. The value of a tag depends on your specific implementation needs.

*The following examples are meant as theoretical examples and may not reflect real-world use cases.*


## Adding tags to a time series

For this example, we are going to use two tags;

- outdoortemp
- GT31

While the first tag is obvious to humans, the second one is the identifier assigned to the sensor from some (theoretical) wiring schematics.

---

1. Navigate to the orange <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/UpdateTimeseriesByUuid">PUT /v2/timeseries/{uuid}</a> field and expand it.

2. Click on the *Try it out* button.

---

3. Erase the example UUID in the field and insert our time series UUID.

4. Erase all fields except "tags" so that you are left with the following;

    ```json
    {
      "tags": [
        "outdoortemp",
        "GT31"
      ]
    }
    ```
5. Click on the blue `Execute` button.

---

6. Scroll down to the response section. As you can see, we got nothing more than an HTTP 204 (No Content) response from the server. This response means that everything is OK and that the server has updated the target.

---

7. If you want to, you can Navigate to the blue <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/FindTimeSeriesByUuid">GET /v2/timeseries/{uuid}</a> field, input the time series UUID, and execute the request to validate that the tags are indeed there.

## Adding tags to things, datasets and programs

The process is the same for the remaining endpoints;

- PUT /v2/things/{uuid}
- PUT /v2/datasets/{uuid}
- PUT /v2/programs/{uuid}

There is little need for us to repeat these steps.

---

Please feel free to play around and modify the tags for the remaining objects we have previously created if you would like to.

---

This section was even shorter!

![On to the next section!][onwards-and-upwards]


# Finding "stuff"

There are several different ways to find what you are looking for.

- Via listing
- Directly via UUIDs
- Via tags

## Listing time series

The following example works for time series, things, datasets and programs.

The interface is the same. The only difference is the specific URLs.

---

1. Navigate to the blue <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/FindTimeSeries">GET /v2/timeseries</a> field and expand it.

2. Click on the *Try it out* button.

---

3. Let all of the fields retain their default value.

4. Click on the blue `Execute` button.

---

5. Scroll down to the response section. The response you get shall look similar to this;

    ```{.json data-line-numbers="|1,15|2-14" data-id="code-animation"}
    [
      {
        "created_by": "00000000-0000-1000-8000-000000000000",
        "lower_bound": -50,
        "name": "My Time Series",
        "si_unit": "C",
        "tags": [
          "outdoortemp",
          "GT31"
        ],
        "thing_uuid": "20b9915d-9208-4a4a-8ab0-6fa45beb8edf",
        "upper_bound": 50,
        "uuid": "6ee20313-67b7-4203-8c02-e882fe454fc3"
      }
    ]
    ```

---

6. Scroll back up and take a look at the limit and offset fields.

    These fields declare the "view" window of the result set.

    - Limit: 20, means that we want (at most) 20 results.
    - Offset: 0, means start at offset 0, or "page" 0 where each page contains "limit" amount of items.

    We only have 1 item, so these parameters do not affect the result at the moment.

---

8. We can use the tags field to affect the result even though we only have one item.
    
9. Click on the "Add string item" button.

10. Change the default value from "string" to "foobar".

11. Click on the blue `Execute` button.

---

12. Scroll down to the response section. The response you get shall look similar to this;

    ```{.json data-line-numbers="" data-id="code-animation"}
    []
    ```

    We found no items matching the query request; No time series has a tag matching "foobar".

13. Go back up and change "foobar" to "outdoortemp". Then execute the request once more.

---

14. Scroll down to the response section. The response you get shall look similar to this;

    ```{.json data-line-numbers="|8|1,15|" data-id="code-animation"}
    [
      {
        "created_by": "00000000-0000-1000-8000-000000000000",
        "lower_bound": -50,
        "name": "My Time Series",
        "si_unit": "C",
        "tags": [
          "outdoortemp",
          "GT31"
        ],
        "thing_uuid": "20b9915d-9208-4a4a-8ab0-6fa45beb8edf",
        "upper_bound": 50,
        "uuid": "6ee20313-67b7-4203-8c02-e882fe454fc3"
      }
    ]
    ```


## Listing things, datasets and programs

As previously mentioned, this works the same way as time series.

Feel free to explore.

---

Ready for the next section? Don't worry; we are done soon.

![On to the next section!][onwards-and-upwards]


# Create a Program

<i class="fas fa-laptop-code"></i>

Let's wire it all together.

---

1. Navigate to the green <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/programs/AddProgram">POST /v2/programs</a> field and expand it.

2. Click on the *Try it out* button.

---

3. Change the request body so that the deadline is 5000 and the schedule is 1m. Giving you a request that looks like this;

    ```{.json data-line-numbers="|2|5|" data-id="code-animation"}
    {
      "deadline": 5000,
      "language": "tengo",
      "name": "My program",
      "schedule": "1m",
      "state": "active",
      "tags": [
        "myprog",
        "awesome"
      ],
      "type": "routine"
    }
    ```

4. Click on the blue `Execute` button.

---

5. Scroll down to the response section. There should be a response looking similar to this, yet with another UUID.

    ```{.json data-line-numbers="|12|" data-id="code-animation"}
    {
      "deadline": 5000,
      "language": "tengo",
      "name": "My program",
      "schedule": "1m",
      "state": "active",
      "tags": [
        "myprog",
        "awesome"
      ],
      "type": "routine",
      "uuid": "80f4501f-e39c-4802-baa9-b2b8be500b55"
    }
    ```

    *Like before, keep track of this UUID.*

## Adding code to a program

Now let's make the program do something.

---

#### But first let's talk about program states

If you look at the API specification, you will notice that a program has three possible states; active, inactive and failed.

At the moment, only active and inactive are used by the Program Manager. An "inactive" state means that the program shall not run, while an "active" state allows the Program Manager to schedule it.

---

When we created our program, we set the initial state to active, yet the program can't do anything without any code to run.

So let's submit that code.

---

#### But before that...

Ensure that the Progam Manager and at least one Program Worker is running.

Good! Now let's submit some code.

---

1. Navigate to the green <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/programs/AddProgramCodeRevision">POST /v2/programs/{uuid}/code</a> field and expand it.

2. Click on the *Try it out* button.

3. Erase the example UUID in the field and insert our program UUID.

---

4. Save the following code to a file called "random_value_insert.tengo".

    Edit it as needed, changing SELFHOST_API_URL and TARGET_TIMESERIES.

    ```go
    base64 := import("base64")
    rand := import("rand")
    json := import("json")
    http := import("http")
    times := import("times")

    SELFHOST_API_URL := "http://127.0.0.1:8095/v2"
    DOMAIN := "test"
    SECRET_KEY := "root"
    TARGET_TIMESERIES := "6ee20313-67b7-4203-8c02-e882fe454fc3"

    store_value := func(shost, timeseries_id, value) {
        headers := {
            "Authorization": "Basic " + base64.encode(DOMAIN + ":" + SECRET_KEY),
            "User-Agent": "Selfhost RandomValueInsert/1.0",
            "Content-Type": "application/json"
        }

        request := [{
            "v": value,
            "ts": times.time_format(times.now(), times.format_rfc3339)
        }]

        response := http.post(
            shost + "/timeseries/" + timeseries_id + "/data",
            [],
            headers,
            json.encode(request)
        )
    }

    rand.seed(times.time_unix_nano(times.now()))
    v := rand.float() * 100 - 50.0

    store_value(SELFHOST_API_URL, TARGET_TIMESERIES, v)
    ```

---

5. Choose the new file by selecting it in the Swagger UI.

6. Click on the blue `Execute` button.

---

7. Scroll down to the response section. There should be a response looking similar to this.

    ```{.json data-line-numbers="|2|5|6,7|" data-id="code-animation"}
    {
      "checksum": "0c0b8a8b9d800fc15184a886eade24bae5a3622081121f1f0e6168bb423d8b26",
      "created": "2021-06-24T13:47:18.869496+02:00",
      "created_by": "00000000-0000-1000-8000-000000000000",
      "revision": 0,
      "signed": null,
      "signed_by": null
    }
    ```

---

While we now have a code (revision 0) assigned to the program. The code is not signed and, as such, will not yet run.

Let's remedy that.


## Signing the code of a program

Requiring the signing of a code revision is meant to avoid having code deployed by the author. While the author may sign the code, this is a bad practice, and the code should be reviewed and then signed by another developer.

---

1. Navigate to the green <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/programs/SignProgramCodeRevisions">POST /v2/programs/{uuid}/revisions/{revision_id}/sign</a> field and expand it.

2. Click on the *Try it out* button.

3. Erase the example UUID in the field and insert our program UUID. Then replace the example revision_id with 0.

4. Click on the blue `Execute` button.

---

5. Scroll down to the response section. As you can see, we got nothing more than an HTTP 204 (No Content) response from the server. This response means that everything is OK and that the server has updated the target.

6. Revision 0 is now signed (by us <i class="fas fa-user-ninja"></i>) and the code will be deployed by the Program Manager.

    *You did start the Program Manager and one Program Worker, right?!*

---


7. To verify this, we can go back and query the time series for data once more.

8. Navigate to the blue <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/timeseries/QueryTimeseriesForData">GET /v2/timeseries/{uuid}/data</a> field and expand it.

9. Click on the *Try it out* button.

---

10. Erase the example UUID in the field and insert our time series UUID.

11. Change the `start` and `end` fields to the period we want to query. A period starting at midnight today and ends 24 hours later will suffice.

12. Leave the remaining input fields on their respective default value.

13. Click on the blue `Execute` button.

---

14. Scroll down to the response section, and there it is.

    ```{.json data-line-numbers="|2-5|6-9|10-13|" data-id="code-animation"}
    [
      {
        "ts": "2021-06-24T13:02:00Z",
        "v": 18.042843
      },
      {
        "ts": "2021-06-24T13:03:00Z",
        "v": -42.58614
      },
      {
        "ts": "2021-06-24T13:04:00Z",
        "v": 16.376854
      }
    ]
    ```

## Modifying the program code

To modify the program, we repeat the steps we already performed;

- Upload a new code revision
- Review it
- Sign it

---

There is no magic to it.

---

Only code.

---

And work.

---

## Tools to aid with software development

As you might have already deduced, the previously mentioned process is not well suited for the day to day process of developing code.

So how do we get around it?

---

#### Use selfctl as a compiler and execute the code locally

The CLI tool, selfctl, which aims to help with everything Self-host related, support the execution of Tengo programs locally.

At the moment, we only target Linux AMD64 (x86_64) and Linux ARM64 targets. In the future, we aim to build binaries for the most common targets; FreeBSD, OpenBSD, Windows and macOS.

---

1. On a Linux machine run;

    > selfctl program compile -f myfile.tengo
    
    To compile a program. Or;

    > selfctrl program run -f myfile.tengo

    To compile and then run a program.

---

Selfctl will report errors in your code and helps with identifying issues before you submit the code as a new revision to a program.

It is not only a helpful tool but an *essential* one.

---

Great work, everybody! We are now on the home stretch. Just a little bit further.

![On to the next section!][onwards-and-upwards]


# Securing the system

So far, we've been using the default access key; in production, We should replace this key as soon as possible to avoid unwanted access to the system.

## Replacing the default access key

This process is relatively straightforward. First, we need to create a new token, store that token someplace secure, then delete the original token.

---

1. Navigate to the green <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/users/AddNewTokenToUser">POST /v2/users/{uuid}/tokens</a> field and expand it.

2. Click on the *Try it out* button.

3. Erase the example UUID in the field and insert the root user UUID, which is always "00000000-0000-1000-8000-000000000000" on a fresh installation.

4. Change the value of "name" in the request body. For this example, we choose;

    ```json
    {
      "name": "New secure token"
    }
    ```

5. Click on the blue `Execute` button.

---

6. Scroll down to the response section. There should be a response looking similar to this.

    ```{.json data-line-numbers="|2|3|4|" data-id="code-animation"}
    {
      "name": "New secret token",
      "secret": "secret-token.nfadcilv7cq2bpa56f43ngqc8wfx6p7tfp7v76e8",
      "uuid": "8366b436-6897-4f89-8835-ed5092830a7f"
    }
    ```

7. Copy the secret and store it someplace safe. Once the insecure token has been removed, this is the only way for you to access the system.

---

8. Click on any of the black padlocks in any of the sections. This action brings up the authorization dialogue.

    Click on logout. Then input "test" as the user and the secret token the server has generated for us.

    Click on authorize, then close the dialogue.

    You are now using the new secure token instead of the old insecure one.

---

9. Navigate to the blue <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/users/FindTokensForUser">GET /v2/users/{uuid}/tokens</a> field and expand it.

10. Click on the *Try it out* button.

11. Erase the example UUID in the field and insert the root user UUID, which is always "00000000-0000-1000-8000-000000000000" on a fresh installation.

12. Click on the blue `Execute` button.

---

13. Scroll down to the response section. There should be a response looking similar to this.

    ```json
    [
      {
        "created": "2021-06-17T07:53:56.622235+02:00",
        "name": "Main admin token",
        "uuid": "ca898044-16d4-4891-9276-c26819e0b9fa"
      },
      {
        "created": "2021-06-24T20:24:50.029749+02:00",
        "name": "New secret token",
        "uuid": "8366b436-6897-4f89-8835-ed5092830a7f"
      }
    ]
    ```
14. Copy the UUID from the old token. The name is "Main admin token".

---

15. Navigate to the blue <a target="_blank" href="http://127.0.0.1:8095/static/swagger-ui#/users/DeleteTokenForUser">GET /v2/users/{uuid}/tokens/{token_uuid}</a> field and expand it.

16. Erase the example UUID in the field and insert the root user UUID, which is always "00000000-0000-1000-8000-000000000000" on a fresh installation. Also, enter the token UUID you just copies.

17. Scroll down to the response section. As you can see, we got nothing more than an HTTP 204 (No Content) response from the server. This response means that everything is OK and that the server has deleted the target.

---

#### The system is now free from insecure tokens.

## Key rotation

It is the process of replacing existing keys with new ones.

---

While the rotation of keys is not implemented in the system itself, one can "easily" implement key rotation if this is a requirement by the organization.

A custom automated solution, backed with a secure store and key distribution solution, can easily rotate the keys as often as you want.

---

## Exposing ports

Never expose a system on a port without requiring authentication to access it.

This rule is often bent on private networks. An unwanted behaviour because anything on a network is a potential target for an infected machine. And for as long as anything on the network can get infected, this problem remains.

In the end, it is better to be safe than sorry.

---

That is it. Thank you for your patience.

![You made it!][use-the-force-link]

# The Code

Everything is available on GitHub;

[https://github.com/self-host/self-host](https://github.com/self-host/self-host)


# Questions?

Now is the time to ask them.


[dangerous-to-go-alone]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/its_dangerous_to_go_alone.png "It's dangerous to go alone"
[you-received-an-item]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/you_received_an_item.png "You received an item!"
[onwards-and-upwards]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/onwards_and_upwards.png "On to the next section!"
[use-the-force-link]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/use_the_force_link.png "Hey, listen!"