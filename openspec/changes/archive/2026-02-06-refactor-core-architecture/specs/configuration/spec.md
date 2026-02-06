# Configuration Specification

## ADDED Requirements

### Requirement: Load Configuration from Environment

The system SHALL load configuration values from environment variables.

**Level**: MUST

#### Scenario: Load configuration with all required values
**Given** all required environment variables are set
**When** loading the configuration
**Then** the system SHALL return a valid configuration struct
**And** all fields SHALL be populated from environment variables
**And** no error SHALL be returned

#### Scenario: Load configuration with default values
**Given** only some environment variables are set
**When** loading the configuration
**Then** the system SHALL return a valid configuration struct
**And** unset variables SHALL use their default values
**And** no error SHALL be returned

#### Scenario: Fail on missing required variable
**Given** a required environment variable is not set
**When** loading the configuration
**Then** the system SHALL return an error
**And** the error SHALL indicate which variable is missing

### Requirement: Validate Configuration Values

The system SHALL validate all loaded configuration values.

**Level**: MUST

#### Scenario: Validate all fields are correct
**Given** configuration is loaded with valid values
**When** validating the configuration
**Then** the validation SHALL succeed
**And** the configuration SHALL be considered valid

#### Scenario: Detect invalid BoltDB path
**Given** the BoltDB path configuration contains invalid characters or path format
**When** validating the configuration
**Then** the validation SHALL fail
**And** the error SHALL specify which field is invalid

#### Scenario: Detect invalid port number
**Given** the HTTP port configuration is outside valid range (1-65535)
**When** validating the configuration
**Then** the validation SHALL fail
**And** the error SHALL indicate the port is out of range

#### Scenario: Detect invalid log level
**Given** the log level configuration is not a valid level (debug, info, warn, error)
**When** validating the configuration
**Then** the validation SHALL fail
**And** the error SHALL indicate the log level is invalid

### Requirement: Provide Configuration Defaults

The system SHALL provide sensible default values for optional configuration fields.

**Level**: MUST

#### Scenario: Use default BoltDB path
**Given** the BOLTDB_PATH environment variable is not set
**When** loading the configuration
**Then** the BoltDB path SHALL default to "./ent.db"
**And** no error SHALL be returned

#### Scenario: Use default HTTP port
**Given** the HTTP_PORT environment variable is not set
**When** loading the configuration
**Then** the HTTP port SHALL default to 8080
**And** no error SHALL be returned

#### Scenario: Use default log level
**Given** the LOG_LEVEL environment variable is not set
**When** loading the configuration
**Then** the log level SHALL default to "info"
**And** no error SHALL be returned

### Requirement: Support Flat Configuration Structure

The configuration structure SHALL be flat with no more than two levels of nesting.

**Level**: MUST

#### Scenario: Access top-level configuration field
**Given** configuration is loaded
**When** accessing a top-level field like BoltDBPath
**Then** the field SHALL be directly accessible without nested struct dereferencing

#### Scenario: All fields are one or two levels deep
**Given** configuration is loaded
**When** examining the configuration structure
**Then** no field SHALL require more than two levels of dereferencing to access
**And** the structure SHALL be simple and readable

### Requirement: Map Environment Variables to Configuration

The system SHALL map environment variable names to configuration fields using clear naming.

**Level**: MUST

#### Scenario: Environment variable name matches field
**Given** environment variable "BOLTDB_PATH" is set
**When** loading the configuration
**Then** the value SHALL be mapped to the BoltDBPath field

#### Scenario: Environment variables use snake_case
**Given** configuration field is BoltDBPath
**When** defining the environment variable
**Then** the variable name SHALL be "BOLTDB_PATH" (snake_case, uppercase)
**And** the mapping SHALL be consistent across all fields

#### Scenario: Environment variable names are discoverable
**Given** a configuration field exists
**When** examining the field definition
**Then** the corresponding environment variable name SHALL be clearly specified in struct tags

### Requirement: Validate BoltDB Configuration

The system SHALL validate BoltDB-specific configuration fields.

**Level**: MUST

#### Scenario: Validate BoltDB file mode
**Given** the BoltDB mode configuration is provided
**When** validating the BoltDB configuration
**Then** the mode SHALL be a valid octal file permission
**And** validation SHALL fail if mode is invalid

#### Scenario: Validate BoltDB timeout
**Given** the BoltDB timeout configuration is provided
**When** validating the BoltDB configuration
**Then** the timeout SHALL be a positive duration
**And** validation SHALL fail if timeout is zero or negative

### Requirement: Validate Server Configuration

The system SHALL validate HTTP server configuration fields.

**Level**: MUST

#### Scenario: Validate HTTP host address
**Given** the HTTP host configuration is provided
**When** validating the server configuration
**Then** the host SHALL be a valid IP address or hostname
**And** validation SHALL fail if host format is invalid

#### Scenario: Validate HTTP port range
**Given** the HTTP port configuration is provided
**When** validating the server configuration
**Then** the port SHALL be between 1 and 65535
**And** validation SHALL fail if port is outside this range

#### Scenario: Validate read timeout
**Given** the HTTP read timeout configuration is provided
**When** validating the server configuration
**Then** the timeout SHALL be a positive duration
**And** validation SHALL fail if timeout is zero or negative

### Requirement: Support Configuration Reload

The system SHALL support reloading configuration without restarting the application.

**Level**: SHOULD

#### Scenario: Reload configuration from environment
**Given** environment variables have changed
**When** reloading the configuration
**Then** the system SHALL reload values from environment variables
**And** the new configuration SHALL be returned
**And** no error SHALL be returned if validation passes

#### Scenario: Validate reloaded configuration
**Given** environment variables have been set to invalid values
**When** attempting to reload the configuration
**Then** the reload SHALL fail with validation error
**And** the previous configuration SHALL remain active

### Requirement: Provide Configuration Help

The system SHALL provide information about available configuration options.

**Level**: SHOULD

#### Scenario: List all configuration options
**When** requesting configuration help
**Then** the system SHALL display all available configuration fields
**And** each field SHALL show its environment variable name
**And** each field SHALL show its default value if applicable
**And** each field SHALL show a brief description

#### Scenario: Show required vs optional fields
**When** requesting configuration help
**Then** the system SHALL indicate which fields are required
**And** the system SHALL indicate which fields have defaults

### Requirement: Handle Configuration Errors Gracefully

The system SHALL provide clear, actionable error messages for configuration failures.

**Level**: MUST

#### Scenario: Error message for missing variable
**Given** a required environment variable is not set
**When** loading the configuration
**Then** the error message SHALL name the missing variable
**And** the error message SHALL suggest how to set it

#### Scenario: Error message for validation failure
**Given** a configuration value fails validation
**When** loading the configuration
**Then** the error message SHALL specify which field failed
**And** the error message SHALL explain why it failed
**And** the error message SHALL suggest valid values or format

#### Scenario: Error message for type conversion failure
**Given** an environment variable cannot be converted to the expected type
**When** loading the configuration
**Then** the error message SHALL indicate the type mismatch
**And** the error message SHALL show the expected type
**And** the error message SHALL show the actual value that failed

### Requirement: Support Configuration Validation Mode

The system SHALL support validating configuration without loading it.

**Level**: SHOULD

#### Scenario: Validate only configuration
**Given** environment variables are set
**When** running in validation mode
**Then** the system SHALL validate all configuration values
**And** the system SHALL report success or failure
**And** the system SHALL not perform any other initialization

#### Scenario: Report all validation errors
**Given** multiple configuration values are invalid
**When** running in validation mode
**Then** the system SHALL report all validation errors
**And** the errors SHALL be clearly separated
**And** the system SHALL exit with failure status
