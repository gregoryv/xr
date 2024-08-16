# Changelog

This project adheres to semantic versioning and all major changes will
be noted in this file.

## [unreleased]

- Remove validation, e.g. minimum, maximum, minLength and maxLength

## [0.8.0] 2024-08-15

- Validate fields after decoding
  e.g. supports field tags `json:"password" minLength:"12"`

## [0.7.0] 2024-08-15

- Check minLength for string fields

## [0.6.0] 2024-08-15

- Check maxLength for string fields
- Check minimum, maximum field tags for number fields

## [0.5.1] 2024-08-14

- Remove logging

## [0.5.0] 2024-08-14

- Add Picker.UseSetter for configurable type setting

## [0.4.0] 2024-08-13

- Ignore body for methods GET, HEAD and DELETE 

## [0.3.0] 2024-07-17

- Add type PickError including source and destination info

## [0.2.0] 2024-07-16

- Require Go 1.22

## [0.1.0] 2024-06-22

- Set values with optional Set method
- Pick comparable types
- Field tags query, path, header and form
