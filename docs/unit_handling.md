# Units (Time Series)

While it makes sense from an academic viewpoint to only use the seven SI base units (second, meter, kilogram, ampere, kelvin, mole and candela) for storing time series data. It does not make as much sense from a performance perspective.

One needs to take the most common use cases into account to avoid introducing rules that will always result in performance penalties.

While it makes sense to store temperature in kelvin as a way to always have a common unit for temperature, it makes less sense in a real world scenario as the kelvin scale is hardly ever used outside of the laboratory.

Having an indoor temperature value on the kelvin scale will always result in convertion to Celsius for all countries except the United States, Belize, Palau, the Bahamas and the Cayman Islands, where Fahrenheit is used.

Most people want to work with value ranges they are confortable with, an indoor temperature of `294.15 K` makes little sense to most people while `21 C` or `80 F` is much clearer.

