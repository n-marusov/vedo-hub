# University Ontology Specification

## Overview
This document defines the university domain ontology for the VEDO platform.

## Classes

### Person
A human being who is part of the university community.
- Subclasses: Student, Professor, StaffMember

### Student
A person enrolled in a course of study.
- Properties: studentId, enrolledIn, gpa
- Relationships: advisedBy (Professor)

### Professor
A faculty member who teaches and conducts research.
- Properties: employeeId, department, researchArea
- Relationships: advises (Student), teaches (Course)

### Course
A unit of teaching that typically lasts one academic term.
- Properties: courseCode, credits, semester
- Relationships: taughtBy (Professor), enrolledBy (Student)

### Department
An academic division within the university.
- Properties: name, budget, headOfDepartment
- Relationships: employs (Professor), offers (Course)

### Organization
An administrative entity.
- Subclasses: Department, ResearchGroup

### ResearchGroup
A team of researchers working on a specific area.
- Properties: focusArea, labLocation
- Relationships: ledBy (Professor), includes (Student)

### Building
A physical structure on campus.
- Properties: address, floorCount, hasWifi

## Properties

### object: advisedBy (Professor, Student)
Inverse of advises. Links a student to their advisor.

### object: advises (Professor, Student)
A professor advises a student.

### object: teaches (Professor, Course)
A professor teaches a course.

### object: enrolledBy (Student, Course)
A student is enrolled in a course.

### object: employs (Department, Professor)
A department employs a professor.

### datatype: studentId (Student, string)
Unique identifier for a student.

### datatype: gpa (Student, float)
Grade point average of a student.

## Individuals

### John (Professor)
Department: Computer Science, Research Area: Ontology Engineering

### Alice (Student)
Enrolled In: CS101, GPA: 3.8
