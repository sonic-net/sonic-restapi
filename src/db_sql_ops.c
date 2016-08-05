/*
 * Copyright (c) 2016 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * THIS CODE IS PROVIDED ON AN  *AS IS* BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
 * LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS
 * FOR A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
 *
 * See the Apache Version 2.0 License for specific language governing
 * permissions and limitations under the License.
 */


/*
 * \file db_sql_ops.c
 * \brief Dell Networking Virtual Machine SQLite Database Operations Wrapper
 */

#include "sonic/db_sql_ops.h"

#include <sonic/std_type_defs.h>
#include <sonic/std_error_codes.h>
#include <sonic/event_log_types.h>
#include <sonic/event_log.h>
#include <sonic/std_utils.h>
#include <sonic/std_assert.h>

#include <stdio.h>
#include <stdlib.h>
#include <sqlite3.h>
#include <string.h>

#define DB_SQL_LOG_TRACE(message, ...) \
    EV_LOGGING(DB_SQL, DEBUG, "DB_SQL", \
        message, ##__VA_ARGS__)

#define DB_SQL_LOG_WARNING(message, ...) \
    EV_LOGGING(DB_SQL, WARNING, "DB_SQL", \
        message, ##__VA_ARGS__)

/**
 * \brief   Construct the sql command
 * \param   [in]  table_name type character string
 * \param   [in]  attribute_name type character string
 * \param   [in]  db_opcode type db_sql_opcode_t
 * \param   [in]  parameter type character string
 * \param   [in]  condition type character string
 * \param   [out] db_sql_cmd type char string
 * \return  t_std_error
 */
static t_std_error db_sql_construct_sql_cmd(const char *table_name,
    db_sql_opcode_t db_opcode, const char *attribute_name,
    const char *parameter, const char *condition, char *db_sql_cmd)
{
    t_std_error rc = STD_ERR_OK;
    switch (db_opcode) {

        case SQL_CREATE:
            snprintf(db_sql_cmd, DB_SQL_CMD_MAX_LEN, " %s %s %s",
                "CREATE TABLE ", table_name, parameter);
            DB_SQL_LOG_TRACE("sql-cmd: %s", db_sql_cmd);
            break;

        case SQL_DROP:
            snprintf(db_sql_cmd, DB_SQL_CMD_MAX_LEN, "%s %s ",
                "DROP TABLE ", table_name);
            DB_SQL_LOG_TRACE("sql-cmd: %s", db_sql_cmd);
            break;

        case SQL_INSERT:
            snprintf(db_sql_cmd, DB_SQL_CMD_MAX_LEN, " %s %s %s %s ",
                "INSERT INTO ", table_name, " values " , parameter);
            DB_SQL_LOG_TRACE("sql-cmd: %s", db_sql_cmd);
            break;

        case SQL_DELETE:
            if (NULL == condition)
                snprintf(db_sql_cmd, DB_SQL_CMD_MAX_LEN, "%s %s ",
                    "DELETE  FROM ", table_name);
            else
                snprintf(db_sql_cmd, DB_SQL_CMD_MAX_LEN, "%s %s %s %s ",
                    "DELETE  FROM ", table_name, " where " , condition);
            DB_SQL_LOG_TRACE("sql-cmd: %s", db_sql_cmd);
            break;

        case SQL_TRIGGER:
            break;

        case SQL_UPDATE:
            snprintf(db_sql_cmd, DB_SQL_CMD_MAX_LEN,
                " %s %s %s %s %s %s %s %s ", "UPDATE", table_name , " set " ,
                attribute_name , " = ", parameter, " where ", condition);
            DB_SQL_LOG_TRACE("sql-cmd: %s", db_sql_cmd);
            break;

        case SQL_SELECT:
            snprintf(db_sql_cmd, DB_SQL_CMD_MAX_LEN,
                "%s %s %s %s %s %s", "SELECT ", attribute_name, " from ",
                table_name , " where " , condition);
            DB_SQL_LOG_TRACE("sql-cmdsql command: %s", db_sql_cmd);
            break;

        default:
            DB_SQL_LOG_WARNING("Error construct SQL command");
            rc = DB_ERR_CONSTRUCT_SQL_CMD;
            break;
    }
    return (rc);
}

/**
 * \brief   Report error condition or info on the Log
 * \param   [in]  rc type t_std_error
 * \param   [in]  log_string type character string
 * \return  t_std_error
 */

static t_std_error db_sql_error_report (t_std_error rc, int  value, const char *log_string)
{
    if (rc != value) {
           DB_SQL_LOG_WARNING("%s failed", log_string);
           return (rc);
    }
    else  {
        DB_SQL_LOG_TRACE("%s success", log_string);
        return(STD_ERR_OK);
    }
}

t_std_error db_sql_open(db_sql_handle_t *ret_handle, const char *db_name)
{
    t_std_error rc = STD_ERR_OK;
    char error_string[DB_SQL_MAX_ERROR_LENGTH] = {0};
    sqlite3 *db_handle;

    STD_ASSERT(db_name != NULL);
    rc = sqlite3_open (db_name, &db_handle);
    sqlite3_exec(db_handle, "PRAGMA foreign_keys = ON;", 0, 0, 0);
    *ret_handle = (db_sql_handle_t) db_handle;

    snprintf(error_string, DB_SQL_MAX_ERROR_LENGTH,
        "db open %s %s", db_name,
        (char *) sqlite3_errmsg((sqlite3 *) db_handle));

    rc = db_sql_error_report (rc, SQLITE_OK, error_string);
    return (rc);
}

t_std_error db_sql_close(db_sql_handle_t db_handle)
{
    t_std_error rc = STD_ERR_OK;
    char error_string[DB_SQL_MAX_ERROR_LENGTH] = {0};

    rc = sqlite3_close((sqlite3 *)db_handle);
    snprintf(error_string, DB_SQL_MAX_ERROR_LENGTH,
        "db close %s ",
        (char *) sqlite3_errmsg((sqlite3 *) db_handle));
    rc = db_sql_error_report (rc, SQLITE_OK, error_string);
    return (rc);
}

t_std_error db_sql_create_table(db_sql_handle_t db_handle,
    const char *table_name, const char *schema)
{
    t_std_error rc = STD_ERR_OK;
    char db_sql_cmd[DB_SQL_CMD_MAX_LEN] = {0};

    STD_ASSERT(table_name != NULL);
    STD_ASSERT(schema != NULL);

    rc = db_sql_construct_sql_cmd (table_name, SQL_CREATE, NULL, schema, NULL, db_sql_cmd);
    if (rc == STD_ERR_OK) {
        rc = db_sql_execute_sql_command(db_handle, db_sql_cmd, SQL_CREATE);
        rc = db_sql_error_report (rc, SQLITE_OK, "CREATE TABLE cmd");
        return (rc);
    }
    else
        return (rc);
}

t_std_error db_sql_drop_table(db_sql_handle_t db_handle, const char *table_name)
{
    t_std_error rc = STD_ERR_OK;
    char db_sql_cmd[DB_SQL_CMD_MAX_LEN] = {0};

    STD_ASSERT(table_name != NULL);

    rc = db_sql_construct_sql_cmd (table_name, SQL_DROP, NULL, NULL, NULL, db_sql_cmd);
    if (rc == STD_ERR_OK) {
        rc = db_sql_execute_sql_command(db_handle, db_sql_cmd, SQL_DROP);
        rc = db_sql_error_report (rc, SQLITE_OK, "DROP TABLE cmd");
        return (rc);
    }
    else
        return (rc);
}

t_std_error db_sql_insert(db_sql_handle_t db_handle, const char *table_name,
    const char *record)
{
    t_std_error rc = STD_ERR_OK;
    char db_sql_cmd[DB_SQL_CMD_MAX_LEN] = {0};

    STD_ASSERT(table_name != NULL);
    STD_ASSERT(record != NULL);

    rc = db_sql_construct_sql_cmd (table_name, SQL_INSERT, NULL, record, NULL, db_sql_cmd);
    if (rc == STD_ERR_OK) {
        rc = db_sql_execute_sql_command(db_handle, db_sql_cmd, SQL_INSERT);
        rc = db_sql_error_report (rc, SQLITE_OK, "INSERT cmd");
        return (rc);
    }
    else
        return (rc);
}

t_std_error db_sql_delete(db_sql_handle_t db_handle, const char *table_name,
    const char *condition)
{
    t_std_error rc = STD_ERR_OK;
    char db_sql_cmd[DB_SQL_CMD_MAX_LEN] = {0};

    STD_ASSERT(table_name != NULL);
    STD_ASSERT(condition != NULL);

    rc = db_sql_construct_sql_cmd (table_name, SQL_DELETE, NULL, NULL, condition, db_sql_cmd);

    if (rc == STD_ERR_OK) {
        rc = db_sql_execute_sql_command(db_handle, db_sql_cmd, SQL_DELETE);
        rc = db_sql_error_report (rc, SQLITE_OK, "DELETE cmd");
        return (rc);
    }
    else
        return (rc);
}

t_std_error db_sql_delete_all_records(db_sql_handle_t db_handle, const char *table_name)
{
    t_std_error rc = STD_ERR_OK;
    char db_sql_cmd[DB_SQL_CMD_MAX_LEN] = {0};

    STD_ASSERT(table_name != NULL);

    rc = db_sql_construct_sql_cmd (table_name, SQL_DELETE, NULL, NULL, NULL, db_sql_cmd);
    if (rc == STD_ERR_OK) {
        rc = db_sql_execute_sql_command(db_handle, db_sql_cmd, SQL_DELETE);

        rc = db_sql_error_report (rc, SQLITE_OK, "DELETE cmd");
        return (rc);
    }
    else
        return(rc);
}


t_std_error db_sql_create_trigger(db_sql_handle_t db_handle, const char *sql_trigger)
{
    t_std_error rc = STD_ERR_OK;

    STD_ASSERT(sql_trigger != NULL);
    DB_SQL_LOG_TRACE("sql command: %s", sql_trigger);

    rc = db_sql_execute_sql_command (db_handle, sql_trigger, SQL_TRIGGER);
    rc = db_sql_error_report (rc, SQLITE_OK, "SQL TRIGGER creation");
    return (rc);
}

/**
 * \brief   Convert sql operation code into a string format
 * \param   [in]  db_opcode type db_sql_opcode_t
 * \param   [out] db_operation type character string
 */

static void db_sql_opcode_to_string (db_sql_opcode_t db_opcode,
    char *db_operation)
{
    switch (db_opcode) {
        case SQL_CREATE:
            snprintf(db_operation, DB_SQL_OPCODE_LENGTH,"%s", "SQL_CREATE");
            break;
        case SQL_DROP:
            snprintf(db_operation, DB_SQL_OPCODE_LENGTH,"%s", "SQL_DROP");
            break;
        case SQL_INSERT:
            snprintf(db_operation, DB_SQL_OPCODE_LENGTH,"%s", "SQL_INSERT");
            break;
        case SQL_DELETE:
            snprintf(db_operation, DB_SQL_OPCODE_LENGTH,"%s", "SQL_DELETE");
            break;
        case SQL_TRIGGER:
            snprintf(db_operation, DB_SQL_OPCODE_LENGTH,"%s", "SQL_TRIGGER");
            break;
        case SQL_UPDATE:
            snprintf(db_operation, DB_SQL_OPCODE_LENGTH,"%s", "SQL_UPDATE");
            break;
        case SQL_SELECT:
            snprintf(db_operation, DB_SQL_OPCODE_LENGTH,"%s", "SQL_SELECT");
            break;
        default:
            snprintf(db_operation, DB_SQL_OPCODE_LENGTH,"%s", "UNKNOWN");
            break;
    }
}

/*
 * Function Prototype definitions for hal db operations
 * Set the attributes of a field in a database table,
 * given the attribute name and condition. Makes a call to
 * the function db_sql_execute_sql_cmd. Returns error code as
 * appropriate
 */

t_std_error db_sql_set_attribute(db_sql_handle_t db_handle, const char *table_name,
    const char *attribute_name, const char *value, const char *condition)
{
    t_std_error rc = STD_ERR_OK;
    char db_sql_cmd[DB_SQL_CMD_MAX_LEN] = {0};

    STD_ASSERT(table_name != NULL);
    STD_ASSERT(attribute_name != NULL);
    STD_ASSERT(value != NULL);
    STD_ASSERT(condition != NULL);

    rc = db_sql_construct_sql_cmd (table_name, SQL_UPDATE, attribute_name, value,
        condition, db_sql_cmd);

    if (rc == STD_ERR_OK) {
        rc = db_sql_execute_sql_command(db_handle, db_sql_cmd, SQL_UPDATE);

        if (rc != SQLITE_OK) {
            DB_SQL_LOG_WARNING("UPDATE cmd failed");
            return (rc);
        }
        else  {
            DB_SQL_LOG_TRACE("UPDATE cmd success");
            return(STD_ERR_OK);
        }
    }
    else
        return(rc);
}

 /* Retrieve attributes from a database table, given the
  * attribute name and query condition. Makes a call to
  * the function db_sql_execute_select_cmd
  * Output is passed in output_str to the caller
  */

t_std_error db_sql_get_attribute(db_sql_handle_t db_handle,
    const char *table_name, const char *attribute_name,
    const char *condition, char *output_str)
{
    t_std_error rc = STD_ERR_OK;
    unsigned int  db_record_count = 0;
    int count = 0;
    char db_output_buffer[DB_SQL_MAX_RECORDS * DB_SQL_OUTPUT_LEN];
    char *db_output[DB_SQL_MAX_RECORDS] = {0};
    char db_sql_cmd[DB_SQL_CMD_MAX_LEN] = {0};

    STD_ASSERT(table_name != NULL);
    STD_ASSERT(attribute_name != NULL);
    STD_ASSERT(condition != NULL);

    rc = db_sql_construct_sql_cmd (table_name, SQL_SELECT, attribute_name, NULL,
        condition, db_sql_cmd);

    if (rc == STD_ERR_OK) {
        for (count = 0; count < DB_SQL_MAX_RECORDS; count++ )
        {
            db_output[count] = &db_output_buffer[count * DB_SQL_OUTPUT_LEN];
            memset (db_output[count], 0, DB_SQL_OUTPUT_LEN);
        }

        rc = db_sql_execute_select_cmd (db_handle, db_sql_cmd,
           &db_record_count, db_output);

        if (rc != SQLITE_OK) {
            DB_SQL_LOG_WARNING("SELECT cmd failed");
            return (rc);
        }
        else  {
            DB_SQL_LOG_TRACE("SELECT cmd success");
            safestrncpy(output_str, db_output[0], DB_SQL_OUTPUT_LEN);
            return(STD_ERR_OK);
        }
    }
    else
        return (rc);
}

/**
 * \brief   perform sqlite finalize
 * \param   [in] db_handle type db_sql_handle_t
 * \param   [in] stmt type pointer to sqlite3_stmt
 * \return  t_std_error
 */

static t_std_error db_sql_finalize(db_sql_handle_t db_handle, sqlite3_stmt *stmt)
{
    t_std_error rc = STD_ERR_OK;

    /* Perofrm the SQL Finalize to cleanup  */

    rc = sqlite3_finalize(stmt);
    if (SQLITE_OK != rc) {
        DB_SQL_LOG_WARNING("ERROR: sqlite finalize %d Reason %s",
            rc, (char *) sqlite3_errmsg(db_handle));
    }
    DB_SQL_LOG_TRACE("sqlite finalize success");
    return (rc);
}

/**
 * For all the sql statements except SELECT,
 * sqlite3_prepare, execute the sqlite statement, perform a sqlite_finalize
 * to do the clean up to avoid any memory leak
 */

t_std_error db_sql_execute_sql_command(db_sql_handle_t db_handle,
    const char *sql_command, db_sql_opcode_t operation_type)
{
    sqlite3_stmt *sql_command_stmt = NULL;
    char db_operation[DB_SQL_OPCODE_LENGTH] = {0};
    t_std_error rc = STD_ERR_OK;

    STD_ASSERT(sql_command != NULL);

    rc = sqlite3_prepare_v2((sqlite3 *) db_handle, sql_command, -1, &sql_command_stmt, NULL);
    if (SQLITE_OK != rc) {
        DB_SQL_LOG_WARNING("ERROR: sqlite preare %d Reason %s",
            rc, (char *) sqlite3_errmsg((sqlite3 *)db_handle));
        return(rc);
    }

    db_sql_opcode_to_string (operation_type, db_operation);

    /* Perofrm the Actual SQL Operation */
    switch (operation_type) {

        case SQL_CREATE:
        case SQL_DROP:
        case SQL_INSERT:
        case SQL_DELETE:
        case SQL_TRIGGER:
        case SQL_UPDATE:

            rc = sqlite3_step(sql_command_stmt);
            if (SQLITE_DONE != rc) {
                DB_SQL_LOG_WARNING("ERROR: sqlite step %d Reason %s",
                    rc, (char *) sqlite3_errmsg((sqlite3 *) db_handle));
            } else {
                DB_SQL_LOG_TRACE("sqlite step success %s",db_operation);
           }
        break;
        default:
            rc = DB_ERR_UNSUPP_OP_TYPE;
            DB_SQL_LOG_WARNING("ERROR: Unsupported operation type %d",
                operation_type);
            return(rc);
    }
    rc = db_sql_finalize(db_handle, sql_command_stmt);
    return(rc);
}

/**
 * For SELECT sql command, perform sqlite3_prepare, execute the SQL
 * statement, perform a sqlite_finalize to do the clean up to avoid
 * any memory leak.
 * Result of SELECT operation, the data values are passed back
 * in the array db_output with a pointer to the record count.
 */

t_std_error db_sql_execute_select_cmd(db_sql_handle_t db_handle,
    const char *sql_command, unsigned int *record_count, char **db_output)
{
    sqlite3_stmt *sql_command_stmt = NULL;
    int rc = 0, column = 0, count = 0;

    STD_ASSERT(sql_command != NULL);

    rc = sqlite3_prepare_v2((sqlite3 *) db_handle, sql_command, -1, &sql_command_stmt, NULL);
    if (SQLITE_OK != rc) {
        DB_SQL_LOG_WARNING("Can't prepare statement. Error %d Reason %s",
            rc, (char *) sqlite3_errmsg((sqlite3 *) db_handle));
        return(rc);
    }

    while(SQLITE_ROW == (rc = sqlite3_step(sql_command_stmt))) {
        for (column = 0; column < sqlite3_column_count(sql_command_stmt);
            column++) {
            safestrncpy (db_output[count],
                (char *) sqlite3_column_text(sql_command_stmt, column),
                DB_SQL_OUTPUT_LEN);
        }
        count++;
    }
    *record_count = count;

    if (SQLITE_DONE != rc) {
            DB_SQL_LOG_WARNING("sql step failed. Error %d Reason %s",
                rc, (char *) sqlite3_errmsg((sqlite3 *) db_handle));
    } else {
        DB_SQL_LOG_TRACE("sql step success");
    }

    rc = db_sql_finalize(db_handle, sql_command_stmt);
    return(rc);
}


t_std_error db_sql_raw_sql_execute (db_sql_handle_t db_handle, const char *input_cmd,
    size_t *output_len, char *output_buff)
{
    sqlite3_stmt *sql_command_stmt = NULL;
    int rc = 0;
    int temp_buf_len = 0;

    STD_ASSERT(input_cmd != NULL);

    rc = sqlite3_prepare_v2((sqlite3 *)db_handle, input_cmd, -1, &sql_command_stmt, NULL);

    if (SQLITE_OK != rc) {
        DB_SQL_LOG_WARNING("sqlite prepare Error %d Reason %s",
            rc, (char *) sqlite3_errmsg((sqlite3 *) db_handle));
        return -1;
    }


    if (NULL == output_buff) {
        rc = sqlite3_step(sql_command_stmt);
        if (SQLITE_DONE != rc) {
            DB_SQL_LOG_WARNING("ERROR: sqlite step %d Reason %s",
                rc, (char *) sqlite3_errmsg((sqlite3 *) db_handle));
        }
    else {
            DB_SQL_LOG_TRACE("sqlite step success");
    }

        rc = db_sql_finalize((sqlite3 *) db_handle, sql_command_stmt);
    }
    else
    {
        while(SQLITE_ROW == (rc = sqlite3_step(sql_command_stmt))) {
            *output_len = sqlite3_column_bytes(sql_command_stmt, 0);
            memcpy(output_buff + temp_buf_len,
                sqlite3_column_text (sql_command_stmt, 0), *output_len);
            temp_buf_len += *output_len;
        }

        if (SQLITE_DONE != rc) {
            DB_SQL_LOG_WARNING("sql step failed. Error %d Reason %s",
                rc, (char *) sqlite3_errmsg((sqlite3 *) db_handle));
        } else
            DB_SQL_LOG_TRACE("sql step success");

        rc = db_sql_finalize(db_handle, sql_command_stmt);

    }

    return rc;
}

