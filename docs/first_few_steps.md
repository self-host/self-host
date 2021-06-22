# First interactions with the Self-hosting API

**Before venturing further**, make sure that you've completed all the necessary steps outlined in [Five to fifteen-minute deployment](https://github.com/self-host/self-host/blob/main/docs/test_deployment.md) walkthrough.

![Dangerous to go alone!][dangerous-to-go-alone]

Throughout this guide, we will use the Swagger UI interface embedded in the Self-host API server. Go ahead and access it now. Authenticate with any user you like, but the user requires full access to the entire system. For simplicity's sake, use the root user if you are using your local development environment.

**You did remember to change the server URL and to add your credentials via the authorize button, right!?**


# Create your first time-series

Let's start with something simple. Let's create a time series to store a temperature value.

1. Scroll down to the `timeseries` section and expand the green `POST /v2/timeseries` field.
2. Click on the *Try it out* button.
3. Start by erasing the `"thing_uuid": "e21ae595-15a5-4f11-8992-9d33600cc1ee",` line. There are no things in the system and indeed nothing with that specific UUID.
4. Then also erase the `tags: [...]` lines as we don't want any tags added to the time series at this point.
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
7. Scroll down below the response example. There should be a response looking similar to this, yet with another UUID.  
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
8. Take note of the UUID in the JSON response, as we will need it in the future.

![You received a new UUID!][you-received-an-item]


### Adding data to a time series

Now, let's add a few data points to our new time series.

1. Scroll down to the `POST /v2/timeseries/{uuid}/data` field and expand it.
2. Click on the *Try it out* button.
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
6. Scroll down below the response example. As you can see, we got nothing more than an HTTP 201 (Created) response from the server. This response means that everything is OK and that the server stored the data in the database.
7. Let's scroll back up and click on the blue `Execute` button again.
8. The response we got this time is not 201 any longer, but 400 with the text: "*Request caused an error due to duplicate key violation*". This error means that there already exists a data point at that particular point in time for this time series. Two data points can occupy the same point in time. If you want to overwrite a data point, you first need to erase it.
9. Let's change the timestamp by adding 1 minute to it. In the example above, we change from `20` to `21`. Then press the blue `Execute` button again.
10. This time, you will again get an HTTP 201 (Created) response from the server.
11. You can repeat step 9 and 10 as many times as you like to add more data points, or you can expand the list of data points and submit more than one at a time, play around with it.

### Retrieve data from a time series

With some data point in our time series, we can now query data from it.

1. Scroll to the blue `GET /v2/timeseries/{uuid}/data` field, located right above the field from the last section and expand it.
2. Click on the *Try it out* button.
3. Erase the example UUID in the field and insert our time series UUID.
4. Change the `start` and `end` fields to the period we want to query. Again, we are keeping in mind that it can not exceed one year or 365.25 days.
5. Leave the remaining input fields on their respective default value.
6. Click on the blue `Execute` button.
7. Scroll down below the response example. There should be a response looking similar to this, yet with different timestamps, possibly different values and numbers of items.
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
8. The query interface supports different operations to apply to the data before creating a response. Such operations include; unit conversion, greater or equal and less or equal filtering, aggregation using precision truncation of the timestamp and time zone declaration.
9. Let's see what happens if we change the requested unit from C(elsius) to F(ahrenheit). Do this by replacing the C in the unit field with an F. Then click on the blue `Execute` button to perform the request once more.
10. Scroll down below the response example. There should be a response looking similar to this, yet with different timestamps, possibly different values and the number of items. Although, as you can see, the value is not the same as before, it is converted from Celsius to Fahrenheit.
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
11. Let's try with something else. Maybe `kWh`. What will happen?
12. What you will get is an HTTP 400 error, with the message "*Unable to convert to the requested unit*" because it is impossible to convert from C to kWh.
13. Change back to C for the unit and set the `precision to **hour** and the `aggregate` to `count`.
14. The result you get depends on the number of data points in the giver period and the range of the period. As we are truncating each timestamp on the hour, then performing a `count` operation on the result, what we are left with is, as might have surmised already, is the number of data points for an hour.

### Delete data from a time series

For the final step in this segment, let's remove one of our data points.

1. Scroll down to the red `DELETE /v2/timeseries/{uuid}/data` field and expand it.
2. Click on the *Try it out* button.
3. Erase the example UUID in the field and insert our time series UUID.
4. Change the `start` and `end` fields to a period so that only a single data point is within it. Leave the `ge` and `le` fields empty.
5. Click on the blue `Execute` button.
6. Scroll down below the response example. As you can see, we got nothing more than an HTTP 204 (Deleted) response from the server. This response means that everything is OK and that the server removed any data in the range from the database.
7. If you want to verify this, you can go back to the previous section and perform a new query over the same period; you shall have one data point less than before.

Great work! On to the next section!

# Creating a Thing

## Connect a time series to a Thing


# Create a dataset

## Updating the dataset

## Retrieving the content of a dataset


# Managing tags

## For a time series

## For a thing

## For a dataset

## For a program


# Finding stuff

## List time series by tags

## List things by tags

## List dataset by tags

## List programs by tags


# Create a Program

## Using selfctl to aid with software development

## Adding code to a program

## Signing the code of a program

## Activating the program

## Modifying a program


[dangerous-to-go-alone]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/its_dangerous_to_go_alone.png "It's dangerous to go alone"
[you-received-an-item]: https://raw.githubusercontent.com/self-host/self-host/main/docs/assets/you_received_an_item.png "You received an item"
